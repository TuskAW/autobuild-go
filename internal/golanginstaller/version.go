package golanginstaller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

// GetLatestGoVersion fetches the latest stable Go version
func GetLatestGoVersion() (string, error) {
	resp, err := http.Get("https://golang.org/dl/?mode=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching Go versions: %v", resp.Status)
	}

	var versions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	for _, version := range versions {
		if stable, ok := version["stable"].(bool); ok && stable {
			versionStr := version["version"].(string)
			return extractVersionNumber(versionStr), nil
		}
	}

	return "", fmt.Errorf("no stable version found")
}

// extractVersionNumber extracts the version number from the string (e.g., "go1.17.2" -> "1.17.2")
func extractVersionNumber(version string) string {
	re := regexp.MustCompile(`go([0-9.]+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
