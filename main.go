package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

// for sorting slice with version
type VersionSlice []*semver.Version

func (s VersionSlice) Len() int           { return len(s) }
func (s VersionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s VersionSlice) Less(i, j int) bool { return s[i].LessThan(*s[j]) }

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	var versionSlice []*semver.Version

	for _, releasedVersion := range releases {
		if minVersion.LessThan(*releasedVersion) {
			// versionSlice = append(versionSlice, releasedVersion)
			versionSlice = insertVersion(versionSlice, releasedVersion)
		}
	}

	sort.Sort(sort.Reverse(VersionSlice(versionSlice)))

	return versionSlice
}

//insert version and make sure the biggest patch of each version is inserted
func insertVersion(currentSlice []*semver.Version, toInsert *semver.Version) []*semver.Version {
	var resultSlice []*semver.Version

	compared := false

	for _, sliceVersion := range currentSlice {
		if sliceVersion.Major != toInsert.Major || sliceVersion.Minor != toInsert.Minor {
			resultSlice = append(resultSlice, sliceVersion)
		} else if toInsert.LessThan(*sliceVersion) {
			resultSlice = append(resultSlice, sliceVersion)
			compared = true
		} else {
			resultSlice = append(resultSlice, toInsert)
			compared = true
		}
	}

	if compared == false {
		resultSlice = append(resultSlice, toInsert)
		return resultSlice
	}

	return resultSlice
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	//open file
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide file name!")
		return
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Github
	client := github.NewClient(nil)
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		stringSlice := strings.Split(scanner.Text(), ",") // the line

		if stringSlice[0] == "repository" && stringSlice[1] == "min_version" {
			continue
		}

		packageSlice := strings.Split(stringSlice[0], "/")

		releases, _, err := client.Repositories.ListReleases(ctx, packageSlice[0], packageSlice[1], opt)
		if err != nil {
			log.Fatal(err)
			// panic(err) // is this really a good way?
		}
		minVersion := semver.New(stringSlice[1])
		allReleases := make([]*semver.Version, len(releases))
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}
			allReleases[i] = semver.New(versionString)
		}
		versionSlice := LatestVersions(allReleases, minVersion)

		fmt.Printf("latest versions of %s: %s\n", stringSlice[0], versionSlice)
	}

}
