package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const repoURL = "https://api.github.com/repos/mateuszmierzwinski/autobuild-go/releases"

type Repo struct {
	TagName string `json:"tag_name"`
	HTMLUrl string `json:"html_url"`
}

var releaseVersion string

// getVersions retrieves the list of releases from the GitHub API
func getVersions() ([]Repo, error) {
	var repos []Repo

	// Fetch data from GitHub API
	result, err := http.Get(repoURL)
	if err != nil {
		return repos, err
	}
	defer result.Body.Close()

	// Check for a successful response
	if result.StatusCode != http.StatusOK {
		return repos, fmt.Errorf("GitHub API error: Status %d", result.StatusCode)
	}

	// Read and unmarshal the response body
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return repos, fmt.Errorf("GitHub API error: Invalid response: %v", err)
	}

	if err = json.Unmarshal(body, &repos); err != nil {
		return repos, fmt.Errorf("GitHub API error: Cannot process response: %v", err)
	}

	return repos, nil
}

// CheckUpdates compares the latest release from GitHub with the current release version
func CheckUpdates() (string, error) {
	allVersions, err := getVersions()
	if err != nil {
		return "", fmt.Errorf("cannot get latest version: %v", err)
	}

	// Ensure there is at least one version in the response
	if len(allVersions) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	// Get the latest release
	latestRelease := allVersions[0]

	// Check if the latest version is newer than the current version
	if latestRelease.TagName != releaseVersion {
		return fmt.Sprintf("New version available: %s. Download here: %s", latestRelease.TagName, latestRelease.HTMLUrl), nil
	}

	return "You are using the latest version.", nil
}
