package main

import (
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

const currentVersion = "0.0.1"

var (
	flagset = flag.NewFlagSet("gh-downloader", flag.ExitOnError)
	flags   = struct {
		GHToken    string
		Repository string
		Asset      string
		Scheme     string
		Output     string
		Version    bool
	}{}
)

func init() {
	flagset.BoolVar(&flags.Version, "version", false, "Print the version and exit")
	flagset.BoolVar(&flags.Version, "v", false, "Print the Version and exit")

	flagset.StringVar(&flags.GHToken, "gh-token", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	flagset.StringVar(&flags.Asset, "a", "", "Asset name")
	flagset.StringVar(&flags.Scheme, "s", "", "Scheme of the release")
	flagset.StringVar(&flags.Output, "o", "", "Output location")

	flagset.StringVar(&flags.Repository, "repository", os.Getenv("REPOSITORY"), "Repository")
}

type Releases []github.RepositoryRelease

func (rs Releases) Len() int      { return len(rs) }
func (rs Releases) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs Releases) Less(i, j int) bool {
	return NewTag(*rs[i].TagName).Less(NewTag(*rs[j].TagName))
}

func DownloadAsset(client *github.Client, asset *github.ReleaseAsset) error {
	req, _ := client.NewRequest("GET", *asset.URL, nil)
	req.Header.Set("Accept", "application/octet-stream")

	apiResponse, err := client.Do(req, nil)

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

func main() {
	flagset.Parse(os.Args[1:])

	if flags.Version {
		fmt.Printf("gh-downloader v%s", currentVersion)
		os.Exit(0)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: flags.GHToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	splittedRepo := strings.Split(flags.Repository, "/")

	if len(splittedRepo) != 2 {
		log.Fatalf("Wrong repository name formatting")
	}

	opt := &github.ListOptions{PerPage: 20}
	var allReleases []github.RepositoryRelease

	for {
		releases, resp, err := client.Repositories.ListReleases(
			splittedRepo[0],
			splittedRepo[1],
			opt,
		)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Release fetching: %s", err.Error()))
		}
		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	filteredReleases := FilterReleases(Releases(allReleases), flags.Scheme)

	if len(filteredReleases) == 0 {
		log.Fatalf("No release matchs to your scheme")
	}

	sort.Sort(filteredReleases)

	release := filteredReleases[len(filteredReleases)-1]

	log.Println(*release.Name)

	for _, asset := range release.Assets {
		if *asset.Name == flags.Asset {
			err := DownloadAsset(client, &asset)

			if err != nil {
				log.Fatalf(err.Error())
			}

			os.Exit(0)
		}
	}

	log.Fatalf("No assets are named like that")
}
