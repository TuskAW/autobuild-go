package config

import (
	"autobuild-go/internal/colors"
	"autobuild-go/internal/models"
	"flag"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

// LoadConfig loads the configuration from a YAML file
func loadConfig(filename string) (*models.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config models.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func parseArg() string {
	profile := flag.String("profile", "default", "Specify the profile to use")
	help := flag.Bool("help", false, "Show this help")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}
	return *profile
}

func GetProfileConfig(projectPath string) models.SelectedConfig {
	profile := parseArg()

	fpath := filepath.Join(projectPath, "autobuild.yaml")
	if _, err := os.Lstat(fpath); err != nil {
		colors.Icon(colors.Yellow, "!!", "No autobuild.yaml in `%s` directory. Using default", projectPath)
		return models.DefaultConfig(projectPath)
	}

	cfg, err := loadConfig(fpath)
	if err != nil {
		colors.Icon(colors.Red, "!!", "Cannot load configuration from `autobuild.yaml` in `%s` directory: %v. Using default", projectPath, err)
		return models.DefaultConfig(projectPath)
	}

	if val, ok := cfg.Profiles[profile]; ok {
		hdir, _ := os.UserHomeDir()
		cfg.Toolchain.Location = strings.Replace(cfg.Toolchain.Location, "$HOME", hdir, -1)
		colors.Success("Profile selected: %s%s%s", colors.Blue, profile, colors.Reset)
		return models.SelectedConfig{
			Profile:   val,
			Toolchain: cfg.Toolchain,
		}
	} else {
		var profiles []string
		for profName, _ := range cfg.Profiles {
			profiles = append(profiles, profName)
		}
		colors.Icon(colors.Red, "!!", "There is no profile named `%s` in `autobuild.yaml` at `%s` directory. Available profiles: %s%s%s", profile, projectPath, colors.Blue, strings.Join(profiles, " "), colors.Reset)
		os.Exit(1)
	}

	return models.DefaultConfig(projectPath)
}
