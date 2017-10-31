package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const currentVersion = "0.0.3"

var (
	flagset = flag.NewFlagSet("gh-downloader", flag.ExitOnError)
	flags   = struct {
		GHToken    string
		Repository string
		Asset      string
		Scheme     string
		Output     string
		Version    bool
		Latest     bool
	}{}
)

func init() {
	flagset.BoolVar(&flags.Version, "version", false, "Print the version and exit")
	flagset.BoolVar(&flags.Version, "v", false, "Print the Version and exit")

	flagset.BoolVar(&flags.Latest, "latest", false, "Fetch the latest release")

	flagset.StringVar(&flags.GHToken, "gh-token", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	flagset.StringVar(&flags.Asset, "a", "", "Asset name")
	flagset.StringVar(&flags.Scheme, "s", "", "Scheme of the release")
	flagset.StringVar(&flags.Output, "o", "", "Output location")

	flagset.StringVar(&flags.Repository, "repository", os.Getenv("REPOSITORY"), "Repository")
}

type releases []*github.RepositoryRelease

func (rs releases) Len() int      { return len(rs) }
func (rs releases) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs releases) Less(i, j int) bool {
	return newTag(*rs[i].TagName).Less(newTag(*rs[j].TagName))
}

func downloadAsset(client *github.Client, asset *github.ReleaseAsset) error {
	req, _ := client.NewRequest("GET", *asset.URL, nil)
	req.Header.Set("Accept", "application/octet-stream")

	apiResponse, err := client.Do(context.Background(), req, nil)

	if err != nil {
		return err
	}

	assetResponse, err := http.Get(apiResponse.Response.Request.URL.String())

	if err != nil {
		return err
	}

	f, err := os.Create(flags.Output)

	if err != nil {
		return err
	}

	defer f.Close()
	io.Copy(f, assetResponse.Body)

	return nil
}

func fetchReleasesByScheme(client *github.Client, org, repo, scheme string) (*github.RepositoryRelease, error) {
	var (
		opt = &github.ListOptions{PerPage: 20}

		allReleases []*github.RepositoryRelease
	)

	for {
		releases, resp, err := client.Repositories.ListReleases(
			context.Background(),
			org,
			repo,
			opt,
		)
		if err != nil {
			return nil, fmt.Errorf("Release fetching: %s", err.Error())
		}
		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	filteredReleases := filterReleases(releases(allReleases), flags.Scheme)

	if len(filteredReleases) == 0 {
		return nil, fmt.Errorf("No release matchs to your scheme")
	}

	sort.Sort(filteredReleases)

	return filteredReleases[len(filteredReleases)-1], nil
}

func fetchLatestRelease(client *github.Client, org, repo string) (*github.RepositoryRelease, error) {
	res, _, err := client.Repositories.GetLatestRelease(context.Background(), org, repo)

	return res, err
}

func main() {
	flagset.Parse(os.Args[1:])

	if flags.Version {
		fmt.Printf("gh-downloader v%s", currentVersion)
		os.Exit(0)
	}

	var (
		splittedRepo = strings.Split(flags.Repository, "/")
		client       = github.NewClient(
			oauth2.NewClient(
				oauth2.NoContext,
				oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: flags.GHToken},
				),
			),
		)

		release *github.RepositoryRelease
		err     error
	)

	if len(splittedRepo) != 2 {
		log.Fatalf("Wrong repository name formatting")
	}

	if flags.Latest {
		release, err = fetchLatestRelease(client, splittedRepo[0], splittedRepo[1])
	} else {
		release, err = fetchReleasesByScheme(
			client,
			splittedRepo[0],
			splittedRepo[1],
			flags.Scheme,
		)
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Release: %s", *release.Name)

	for _, asset := range release.Assets {
		if *asset.Name == flags.Asset {
			err := downloadAsset(client, &asset)

			if err != nil {
				log.Fatalf(err.Error())
			}

			os.Exit(0)
		}
	}

	log.Fatalf("No assets are named like that")
}
