package main

import (
	"fmt"
	"strings"
)

type tag struct {
	Project string
	Major   int
	Minor   int
	Patch   int
}

func newTag(name string) *tag {
	var (
		t tag

		version = name
	)

	if splittedTagName := strings.Split(name, "-"); len(splittedTagName) > 1 {
		t.Project = strings.Join(splittedTagName[:len(splittedTagName)-1], "-")
		version = splittedTagName[len(splittedTagName)-1]
	}

	if n, err := fmt.Sscanf(
		strings.TrimPrefix(version, "v"),
		"%d.%d.%d",
		&t.Major,
		&t.Minor,
		&t.Patch,
	); n != 3 || err != nil {
		return nil
	}

	return &t
}

func (t1 *tag) Less(t2 *tag) bool {
	if t1.Project != t2.Project {
		return false
	}

	return t1.Major < t2.Major ||
		(t1.Major == t2.Major && t1.Minor < t2.Minor) ||
		(t1.Major == t2.Major && t1.Minor == t2.Minor && t1.Patch < t2.Patch)
}

func filterReleases(rs releases, scheme string) releases {
	var filteredReleases releases
	schemeTag := newTag(scheme)

	if schemeTag == nil {
		return filteredReleases
	}

	for _, release := range rs {
		r := newTag(*release.TagName)

		if r == nil {
			continue
		}

		if r.Project == schemeTag.Project &&
			(schemeTag.Major == -1 || schemeTag.Major == r.Major) &&
			(schemeTag.Minor == -1 || schemeTag.Minor == r.Minor) &&
			(schemeTag.Patch == -1 || schemeTag.Patch == r.Patch) {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases
}
