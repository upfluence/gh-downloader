package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/upfluence/cfg/x/cli"
	"golang.org/x/oauth2"
)

const assetFmt = "{{ .repo }}-{{ .os }}-{{ .arch }}-{{ .version }}"

var defaultConfig = config{
	FileMode: 0644,
	Asset: templateConfig{
		t:   template.Must(template.New("").Parse(assetFmt)),
		fmt: assetFmt,
	},
}

type fileMode os.FileMode

func (fm *fileMode) Parse(v string) error {
	_, err := fmt.Sscanf(v, "%o", fm)
	return err
}

func (fm fileMode) String() string { return fmt.Sprintf("%04o", fm) }

type repoConfig struct {
	owner string
	repo  string
}

func (rc repoConfig) isZero() bool { return rc == repoConfig{} }

func (rc *repoConfig) Parse(v string) error {
	var parts = strings.Split(v, "/")

	switch len(parts) {
	case 1:
		rc.owner = "upfluence"
		rc.repo = parts[0]
	case 2:
		rc.owner = parts[0]
		rc.repo = parts[1]
	default:
		return fmt.Errorf("Invalid repo format: %q", v)
	}

	return nil
}

func (rc repoConfig) String() string {
	return fmt.Sprintf("%s/%s", rc.owner, rc.repo)
}

type templateConfig struct {
	t   *template.Template
	fmt string
}

func (tc *templateConfig) Parse(v string) error {
	t, err := template.New("").Parse(v)

	if err == nil {
		tc.t = t
		tc.fmt = v
	}

	return err
}

func (tc templateConfig) String() string { return tc.fmt }

type config struct {
	GithubToken string         `env:"GITHUB_TOKEN,UPF_GITHUB_TOKEN" flag:"gh-token" help:"GitHub API token"`
	Repository  repoConfig     `flag:"repository,r" help:"Repository"`
	Asset       templateConfig `flag:"asset,a" help:"Asset name, it can be given as a go template with variable including: .tag, .repo, .arch, .os, .version"`
	Scheme      string         `flag:"scheme,s" help:"Scheme of the release, if empty the latest release will be pulled"`
	Output      string         `flag:"output,o" help:"Output location on disk, if left empty the file will be written on stdout"`
	FileMode    fileMode       `flag:"mode" help:"File mode to create the output file in"`
}

func (c *config) client(ctx context.Context) *github.Client {
	return github.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.GithubToken}),
		),
	)
}

func (c *config) output(cctx cli.CommandContext) (io.WriteCloser, error) {
	if c.Output == "" {
		return nopCloser{Writer: cctx.Stdout}, nil
	}

	return os.OpenFile(
		c.Output,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		os.FileMode(c.FileMode),
	)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func main() {
	cli.NewApp(
		cli.WithName("gh-downloader"),
		cli.WithCommand(
			cli.StaticCommand{
				Help:     cli.HelpWriter(&defaultConfig),
				Synopsis: cli.SynopsisWriter(&defaultConfig),
				Execute: func(ctx context.Context, cctx cli.CommandContext) error {
					var (
						release *github.RepositoryRelease

						c = defaultConfig
					)

					if err := cctx.Configurator.Populate(ctx, &c); err != nil {
						return err
					}

					repo := c.Repository

					if repo.isZero() {
						return errors.New("no repository provided")
					}

					w, err := c.output(cctx)

					if err != nil {
						return err
					}

					defer w.Close()

					cl := c.client(ctx)

					if scheme := strings.TrimSpace(c.Scheme); scheme == "" {
						fmt.Fprintf(cctx.Stdout, "Fetching latest release of %v\n", repo)
						release, err = fetchLatestRelease(cl, repo.owner, repo.repo)
					} else {
						fmt.Fprintf(
							cctx.Stdout,
							"Fetching release of %v matching scheme %q\n",
							repo,
							scheme,
						)
						release, err = fetchReleasesByScheme(cl, repo.owner, repo.repo, scheme)
					}

					if err != nil {
						return err
					}

					assetName, err := templateAssetName(&c, release)

					if err != nil {
						return err
					}

					fmt.Fprintf(cctx.Stdout, "Release: %s [%s]\n", *release.Name, assetName)

					for _, asset := range release.Assets {
						if *asset.Name == assetName {
							return downloadAsset(w, cl, repo.owner, repo.repo, &asset)
						}
					}

					return fmt.Errorf("No assets are named %q", assetName)
				},
			},
		),
	).Run(context.Background())
}

func templateAssetName(c *config, r *github.RepositoryRelease) (string, error) {
	var buf bytes.Buffer

	if err := c.Asset.t.Execute(
		&buf,
		map[string]string{
			"repo":    c.Repository.repo,
			"owner":   c.Repository.owner,
			"tag":     *r.TagName,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
			"version": strings.TrimPrefix(*r.TagName, "v"),
		},
	); err != nil {
		return "", err
	}

	return buf.String(), nil
}
