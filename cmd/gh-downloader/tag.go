package main

import (
	"regexp"
	"strconv"
	"strings"
)

var versionRe = regexp.MustCompile(`((\w+)-)?v?(\d+|x).(\d+|x).(\d+|x)(-(\w+))?`)

type tag struct {
	Project string
	Major   int
	Minor   int
	Patch   int
	Suffix  string
}

func newTag(name string) *tag {
	var ms = versionRe.FindStringSubmatch(name)

	if len(ms) != 8 {
		return nil
	}

	return &tag{
		Project: ms[2],
		Major:   parseVersionNumber(ms[3]),
		Minor:   parseVersionNumber(ms[4]),
		Patch:   parseVersionNumber(ms[5]),
		Suffix:  ms[7],
	}
}

func parseVersionNumber(sv string) int {
	if v, err := strconv.Atoi(sv); err == nil {
		return v
	}

	return -1
}

func (t1 *tag) Less(t2 *tag) bool {
	if t1.Project != t2.Project {
		return strings.Compare(t1.Project, t2.Project) < 0
	}

	if t1.Major != t2.Major {
		return t1.Major < t2.Major
	}

	if t1.Minor != t2.Minor {
		return t1.Minor < t2.Minor
	}

	if t1.Patch != t2.Patch {
		return t1.Patch < t2.Patch
	}

	if t1.Suffix == "" || t2.Suffix == "" {
		return t2.Suffix == ""
	}

	return strings.Compare(t1.Suffix, t2.Suffix) < 0
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
			(schemeTag.Patch == -1 || schemeTag.Patch == r.Patch) &&
			(schemeTag.Suffix == "*" || schemeTag.Suffix == r.Suffix) {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases
}
