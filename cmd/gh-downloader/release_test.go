package main

import (
	"sort"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

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
				out[i] = *rr.TagName
			}

			assert.Equal(t, tt.want, out)
		})
	}
}

func newRelease(tag string) github.RepositoryRelease {
	return github.RepositoryRelease{
		TagName: &tag,
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
