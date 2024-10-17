package processors

import (
	"autobuild-go/internal/colors"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

import (
	"autobuild-go/internal/models"
	"sync"
)

// ProjectWalker is responsible for scanning the source tree and finding main.go and its corresponding go.mod
type ProjectWalker struct {
	path        string
	buildDir    string
	projectDest chan models.Project
}

func (p *ProjectWalker) Run() error {
	if p.path == "" || p.projectDest == nil {
		return errors.New("projectWalker not initialized")
	}
	return p.processPath(p.path, p.projectDest)
}

func (p *ProjectWalker) processPath(path string, dest chan models.Project) error {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		findMainAndGoMod(path, p.buildDir, dest)
	}()

	wg.Wait()
	close(dest)
	return nil
}

// findMainAndGoMod scans directories to find main.go and pair it with its nearest go.mod
func findMainAndGoMod(startPath string, buildDir string, dest chan models.Project) {
	filepath.WalkDir(startPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			colors.ErrLog("Error accessing path %s: %v", path, err)
			return nil
		}

		if strings.Contains(path, ".toolchain") || strings.Contains(path, ".build") {
			return nil
		}

		// Check if it's a file and if it's named "main.go"
		if !d.IsDir() && strings.ToLower(d.Name()) == "main.go" {
			// We have found a main.go, now let's search for the closest go.mod upwards
			mainGoDir := filepath.Dir(path)
			goModDir := findNearestGoMod(mainGoDir)

			if goModDir != "" {
				// We found a valid go.mod, construct a Project object
				appName := filepath.Base(mainGoDir)
				project := models.Project{
					AppName:       appName,
					AppMainSrcDir: mainGoDir,
					RootDir:       goModDir,
					BuildDir:      buildDir,
				}
				// Send the constructed project to the channel
				dest <- project
			}
		}
		return nil
	})
}

// findNearestGoMod searches for the closest go.mod starting from the given directory and moving upwards
func findNearestGoMod(startDir string) string {
	currentDir := startDir
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// go.mod found in this directory
			return currentDir
		}

		// Move up a directory level
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We have reached the root, no go.mod found
			return ""
		}
		currentDir = parentDir
	}
}

// NewProjectWalkerProcessor constructs a ProjectWalker and returns it as a Processor
func NewProjectWalkerProcessor(path, buildDir string, projectDest chan models.Project) Processor {
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		os.MkdirAll(buildDir, os.ModePerm)
	}

	return &ProjectWalker{
		path:        path,
		buildDir:    buildDir,
		projectDest: projectDest,
	}
}
