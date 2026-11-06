package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/google/go-github/v93/github"
)

type releases []*github.RepositoryRelease

func (rs releases) Len() int      { return len(rs) }
func (rs releases) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs releases) Less(i, j int) bool {
	return newTag(rs[i].TagName).Less(newTag(rs[j].TagName))
}

type releaseClient struct {
	cl    *github.Client
	owner string
	repo  string
}

func (rc *releaseClient) openAsset(ctx context.Context, asset *github.ReleaseAsset) (io.ReadCloser, error) {
	r, _, err := rc.cl.Repositories.DownloadReleaseAsset(
		ctx,
		rc.owner,
		rc.repo,
		asset.GetID(),
		http.DefaultClient,
	)

	if err != nil {
		return nil, fmt.Errorf("download release asset: %w", err)
	}

	return r, nil
}

func (rc *releaseClient) releaseByScheme(ctx context.Context, scheme string) (*github.RepositoryRelease, error) {
	var (
		opt = &github.ListOptions{PerPage: 100}

		allReleases []*github.RepositoryRelease
	)

	for {
		rs, resp, err := rc.cl.Repositories.ListReleases(
			ctx,
			rc.owner,
			rc.repo,
			opt,
		)

		if err != nil {
			return nil, fmt.Errorf("list releases: %w", err)
		}

		allReleases = append(allReleases, rs...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	filteredReleases := filterReleases(releases(allReleases), scheme)

	if len(filteredReleases) == 0 {
		return nil, fmt.Errorf(
			"no release matches the scheme %q out of %d releases",
			scheme,
			len(allReleases),
		)
	}

	sort.Sort(filteredReleases)

	return filteredReleases[len(filteredReleases)-1], nil
}

func (rc *releaseClient) latestRelease(ctx context.Context) (*github.RepositoryRelease, error) {
	r, _, err := rc.cl.Repositories.GetLatestRelease(ctx, rc.owner, rc.repo)

	if err != nil {
		return nil, fmt.Errorf("get latest release: %w", err)
	}

	return r, nil
}
