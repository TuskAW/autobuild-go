package models

import "path/filepath"

// Profile represents the structure of each profile in the YAML
type Profile struct {
	OS     map[string][]string `yaml:"os"`     // Operating systems with architectures
	Stages []string            `yaml:"stages"` // List of stages
}

// Toolchain represents the static toolchain configuration
type Toolchain struct {
	Golang   string `yaml:"golang"`
	Location string `yaml:"location"`
}

// Config is the main structure containing profiles and toolchain
type Config struct {
	Profiles  map[string]Profile `yaml:"profiles"`  // Map of profiles for easy selection by name
	Toolchain Toolchain          `yaml:"toolchain"` // Toolchain configuration
}

type SelectedConfig struct {
	Profile        Profile
	Toolchain      Toolchain
	CurrentVersion string
}

func DefaultConfig(path string) SelectedConfig {
	return SelectedConfig{
		Profile: Profile{
			OS: map[string][]string{
				"windows": {"amd64", "arm64"},
				"linux":   {"amd64", "arm64"},
				"darwin":  {"amd64", "arm64"},
			},
			Stages: []string{
				"test", "build", "hash",
			},
		},
		Toolchain: Toolchain{
			Golang:   "latest",
			Location: filepath.Join(filepath.Dir(path), ".toolchain"),
		},
	}
}
