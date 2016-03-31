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
	splittedTagName := strings.Split(name, "-")
	project := ""
	version := ""

	if len(splittedTagName) == 1 {
		version = splittedTagName[0]
	} else if len(splittedTagName) == 2 {
		version = splittedTagName[1]
	} else {
		return nil
	}

	if version[0:1] == "v" {
		version = version[1:]
	}

	splittedVersion := strings.Split(version, ".")

	if len(splittedVersion) != 3 {
		return nil
	}

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

	return &Tag{project, major, minor, patch}
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

	if schemeTag == nil {
		return filteredReleases
	}

	for _, release := range releases {
		r := NewTag(*release.TagName)

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
