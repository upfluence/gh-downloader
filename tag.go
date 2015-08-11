package main

import (
	"strconv"
	"strings"
)

type Tag struct {
	Project string
	Major   int
	Minor   int
	Patch   int
}

func NewTag(name string) *Tag {
	splittedTagName := strings.Split(name, "-v")
	splittedVersion := strings.Split(splittedTagName[1], ".")

	patch, err := strconv.Atoi(splittedVersion[2])
	if err != nil {
		patch = -1
	}

	minor, err := strconv.Atoi(splittedVersion[1])
	if err != nil {
		minor = -1
	}

	major, err := strconv.Atoi(splittedVersion[0])
	if err != nil {
		major = -1
	}

	return &Tag{splittedTagName[0], major, minor, patch}
}

func (t1 *Tag) Less(t2 *Tag) bool {
	if t1.Project != t2.Project {
		return false
	}

	return t1.Major < t2.Major ||
		(t1.Major == t2.Major && t1.Minor < t2.Minor) ||
		(t1.Major == t2.Major && t1.Minor == t2.Minor && t1.Patch < t2.Patch)
}

func FilterReleases(releases Releases, scheme string) Releases {
	filteredReleases := Releases{}
	schemeTag := NewTag(scheme)

	for _, release := range releases {
		r := NewTag(*release.TagName)

		if r.Project == schemeTag.Project &&
			(schemeTag.Major == -1 || schemeTag.Major == r.Major) &&
			(schemeTag.Minor == -1 || schemeTag.Minor == r.Minor) &&
			(schemeTag.Patch == -1 || schemeTag.Patch == r.Patch) {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases
}
