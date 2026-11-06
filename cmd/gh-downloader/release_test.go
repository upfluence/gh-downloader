package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const assetContent = "binary"

func TestReleasesSort(t *testing.T) {
	for _, tt := range []struct {
		name string
		with releases
		want []string
	}{
		{
			name: "path",
			with: buildReleases("foo-1.2.100", "foo-1.2.2", "foo-1.2.30"),
			want: []string{"foo-1.2.2", "foo-1.2.30", "foo-1.2.100"},
		},
		{
			name: "minor",
			with: buildReleases("foo-1.200.2", "foo-1.1.2", "foo-1.30.2"),
			want: []string{"foo-1.1.2", "foo-1.30.2", "foo-1.200.2"},
		},
		{
			name: "major",
			with: buildReleases("foo-300.2.1", "foo-2.2.1", "foo-50.2.1"),
			want: []string{"foo-2.2.1", "foo-50.2.1", "foo-300.2.1"},
		},
		{
			name: "with suffix",
			with: buildReleases("foo-1.2.2-rc19", "foo-1.2.2-rc10", "foo-1.2.2-rc1"),
			want: []string{"foo-1.2.2-rc1", "foo-1.2.2-rc10", "foo-1.2.2-rc19"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.with)

			var out = make([]string, len(tt.with))

			for i, rr := range tt.with {
				out[i] = rr.TagName
			}

			assert.Equal(t, tt.want, out)
		})
	}
}

func TestOpenAsset(t *testing.T) {
	for _, tt := range []struct {
		name       string
		haveStatus int
		want       string
		assertErr  assert.ErrorAssertionFunc
	}{
		{
			name:       "redirected asset",
			haveStatus: http.StatusOK,
			want:       assetContent,
			assertErr:  assert.NoError,
		},
		{
			name:       "redirected asset error page",
			haveStatus: http.StatusForbidden,
			assertErr:  assert.Error,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			storage := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(tt.haveStatus)
					io.WriteString(w, assetContent)
				}),
			)

			defer storage.Close()

			api := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, storage.URL+"/asset", http.StatusFound)
				}),
			)

			defer api.Close()

			cl, err := github.NewClient(
				github.WithURLs(new(api.URL+"/"), nil),
			)

			require.NoError(t, err)

			relc := releaseClient{cl: cl, owner: "upfluence", repo: "ucheck"}

			rc, err := relc.openAsset(
				t.Context(),
				&github.ReleaseAsset{ID: new(int64(1))},
			)

			tt.assertErr(t, err)

			var got string

			if rc != nil {
				buf, err := io.ReadAll(rc)

				require.NoError(t, err)

				rc.Close()

				got = string(buf)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func newRelease(tag string) github.RepositoryRelease {
	return github.RepositoryRelease{
		TagName: tag,
	}
}

func buildReleases(tags ...string) releases {
	var out releases

	for _, t := range tags {
		r := newRelease(t)
		out = append(out, &r)
	}

	return out
}
