package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"

	"github.com/google/go-github/v93/github"
	"github.com/upfluence/cfg/x/cli"
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
	IfNotExist  bool           `flag:"if-not-exist" env:"IF_NOT_EXIST"`
}

func (c *config) releaseClient() (*releaseClient, error) {
	var opts []github.ClientOptionsFunc

	if c.GithubToken != "" {
		opts = append(opts, github.WithAuthToken(c.GithubToken))
	}

	cl, err := github.NewClient(opts...)

	if err != nil {
		return nil, fmt.Errorf("build github client: %w", err)
	}

	return &releaseClient{
		cl:    cl,
		owner: c.Repository.owner,
		repo:  c.Repository.repo,
	}, nil
}

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

					if c.IfNotExist {
						_, err := os.Stat(c.Output)

						switch {
						case err == nil:
							return nil
						case os.IsNotExist(err):
						default:
							return err
						}
					}

					cl, err := c.releaseClient()

					if err != nil {
						return err
					}

					if scheme := strings.TrimSpace(c.Scheme); scheme == "" || scheme == "latest" {
						fmt.Fprintf(cctx.Stderr, "Fetching latest release of %v\n", repo)

						release, err = cl.latestRelease(ctx)
					} else {
						fmt.Fprintf(
							cctx.Stderr,
							"Fetching release of %v matching scheme %q\n",
							repo,
							scheme,
						)

						release, err = cl.releaseByScheme(ctx, scheme)
					}

					if err != nil {
						return err
					}

					assetName, err := templateAssetName(&c, release)

					if err != nil {
						return err
					}

					fmt.Fprintf(cctx.Stderr, "Release: %s [%s]\n", release.GetName(), assetName)

					for _, asset := range release.Assets {
						if asset.GetName() == assetName {
							return downloadAsset(
								ctx,
								cl,
								asset,
								c.Output,
								os.FileMode(c.FileMode),
								cctx.Stdout,
							)
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
			"tag":     r.TagName,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
			"version": strings.TrimPrefix(r.TagName, "v"),
		},
	); err != nil {
		return "", err
	}

	return buf.String(), nil
}
