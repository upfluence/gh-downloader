package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/google/go-github/github"
)

type releases []*github.RepositoryRelease

func (rs releases) Len() int      { return len(rs) }
func (rs releases) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs releases) Less(i, j int) bool {
	return newTag(*rs[i].TagName).Less(newTag(*rs[j].TagName))
}

func downloadAsset(w io.Writer, client *github.Client, org, repo string, asset *github.ReleaseAsset) error {
	buf, loc, err := client.Repositories.DownloadReleaseAsset(
		context.Background(),
		org,
		repo,
		*asset.ID,
	)

	if err != nil {
		return err
	}

	if buf == nil {
		resp, err := http.Get(loc)

		if err != nil {
			return err
		}

		buf = resp.Body
	}

	defer buf.Close()

	_, err = io.Copy(w, buf)
	return err
}

func fetchReleasesByScheme(client *github.Client, org, repo, scheme string) (*github.RepositoryRelease, error) {
	var (
		opt = &github.ListOptions{PerPage: 100}

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

	filteredReleases := filterReleases(releases(allReleases), scheme)

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
