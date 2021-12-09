package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTag(t *testing.T) {
	for _, tt := range []struct {
		with string
		want tag
	}{
		{
			with: "foo-x.x.x",
			want: tag{Project: "foo", Major: -1, Minor: -1, Patch: -1, Suffix: ""},
		},
		{
			with: "foo-vx.x.x",
			want: tag{Project: "foo", Major: -1, Minor: -1, Patch: -1, Suffix: ""},
		},
		{
			with: "foo-1.2.3",
			want: tag{Project: "foo", Major: 1, Minor: 2, Patch: 3, Suffix: ""},
		},
		{
			with: "foo-v1.2.3",
			want: tag{Project: "foo", Major: 1, Minor: 2, Patch: 3, Suffix: ""},
		},
		{
			with: "foo-10.20.30",
			want: tag{Project: "foo", Major: 10, Minor: 20, Patch: 30, Suffix: ""},
		},
		{
			with: "foo-10.20.30-bar",
			want: tag{Project: "foo", Major: 10, Minor: 20, Patch: 30, Suffix: "bar"},
		},
	} {
		t.Run(tt.with, func(t *testing.T) {
			tag := newTag(tt.with)

			assert.Equal(t, tt.want, *tag)
		})
	}
}
