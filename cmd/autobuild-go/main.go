package main

import (
	"autobuild-go/internal/builder"
	"autobuild-go/internal/colors"
	"autobuild-go/internal/config"
	"autobuild-go/internal/golanginstaller"
	"autobuild-go/internal/models"
	"autobuild-go/internal/processors"
	"fmt"
	_ "gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const autobuildGoHeader = "    _       _       ___      _ _    _      ___     \n   /_\\ _  _| |_ ___| _ )_  _(_) |__| |___ / __|___ \n  / _ \\ || |  _/ _ \\ _ \\ || | | / _` |___| (_ / _ \\\n /_/ \\_\\_,_|\\__\\___/___/\\_,_|_|_\\__,_|    \\___\\___/\n                                                   "

// Check if git is installed
func isGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func main() {

	fmt.Printf(colors.Purple+autobuildGoHeader+colors.Reset+"\n\t%d (c) Mateusz Mierzwinski - matt@mattmierzwinski.com\n\tThis is a free software released under BSD-2 simplified license.\n\tSource: https://github.com/mateuszmierzwinski/autobuild-go\n\n", time.Now().Year())

	colors.HorizontalLine("Environment check")
	if !isGitInstalled() {
		colors.ErrLog("git is not installed. Please install git and try again.")
		return
	}
	colors.Success("Git is installed.")

	var path string

	deflen := 1
	allFlag := strings.Join(os.Args, " ")
	switch {
	case strings.Contains(allFlag, "--profile"):
		deflen += 2
	case strings.Contains(allFlag, "--help"):
		deflen += 1
	}

	if len(os.Args) > deflen {
		path = os.Args[len(os.Args)-1]
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			colors.ErrLog("Error trying to get current dir: %v", err)
			os.Exit(1)
		}
	}

	conf := config.GetProfileConfig(path)

	// Create a new GoInstaller instance
	installer := golanginstaller.New(path, conf)

	// Ensure Go is installed
	if err := installer.EnsureGo(); err != nil {
		colors.ErrLog("Error ensuring Go is installed: %v", err)
		os.Exit(1)
	}

	colors.HorizontalLine("Testing & building Go projects")

	projectDestChan := make(chan models.Project, 5)

	proc := processors.NewProjectWalkerProcessor(path, filepath.Join(path, ".build"), projectDestChan)
	gobuilder := builder.NewGoBuilder(installer.GoToolchainDir(), conf)

	// Run the processor in a separate goroutine
	go func() {
		if err := proc.Run(); err != nil {
			colors.ErrLog("Error running project walker: %v", err)
			os.Exit(1)
		}
	}()

	gobuilder.Build(projectDestChan)

	colors.HorizontalLine("Done!")
}
