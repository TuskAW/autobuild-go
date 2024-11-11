package builder

import (
	"autobuild-go/internal/colors"
	"autobuild-go/internal/models"
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type GoBuilderTargets []GoBuilderTarget

type GoBuilderTarget struct {
	GOOS       string
	GOARCH     string
	EXECSUFFIX string
}

type GoBuilder struct {
	toolchainPath string
	goRootPath    string
	goPathPath    string
	targets       GoBuilderTargets
	defaultEnv    []string
	hashers       map[string]hash.Hash
	stages        map[string]bool
}

func (g *GoBuilder) Build(projectsSource chan models.Project) {
	g.defaultEnv = append(g.defaultEnv, fmt.Sprintf("GOPATH=%s", g.goPathPath))
	g.defaultEnv = append(g.defaultEnv, fmt.Sprintf("GOROOT=%s", g.goRootPath))
	g.defaultEnv = append(g.defaultEnv, fmt.Sprintf("PATH=%s%c%s", filepath.Join(g.goRootPath, "bin"), os.PathListSeparator, os.Getenv("PATH")))

	wg := sync.WaitGroup{}
	for project := range projectsSource {
		wg.Add(1)
		go func(project models.Project) {
			defer wg.Done()
			if _, ok := g.stages["test"]; ok {
				if err := g.testExec(project); err != nil {
					colors.ErrLog("Error: %v", err)
					return
				}
			}

			if _, ok := g.stages["build"]; ok {
				if err := g.buildExec(project); err != nil {
					colors.ErrLog("Error building app %s: %v", project.AppName, err)
					return
				}
			}

			if _, ok := g.stages["hash"]; ok {
				if err := g.sumExec(project); err != nil {
					colors.ErrLog("Error creating SHA-256 sum for app %s: %v", project.AppName, err)
					return
				}
			}
		}(project)
	}
	wg.Wait()
}

func (g *GoBuilder) testExec(project models.Project) error {
	tn := time.Now()
	colors.Icon(colors.Yellow, "\u226b", "Testing app "+colors.Blue+"%s"+colors.Reset, project.AppName)

	// Prepare the build command: go build -o outputPath project.AppMainSrcDir
	cmd := exec.Command(filepath.Join(g.goRootPath, "bin", "go"), "test", "-v", fmt.Sprintf("-coverprofile=%s/coverage-%s.txt", project.BuildDir, project.AppName), "./...")
	cmd.Dir = project.RootDir
	cmd.Env = g.defaultEnv

	// Capture output
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Execute the command
	if err := cmd.Run(); err != nil {
		// If there's an error, return the captured stdout and stderr as part of the error
		persistLog(outBuf, errBuf, project.BuildDir, project.AppName)
		return fmt.Errorf("error testing %s: %v. Logs created", project.AppName, err)
	}
	colors.Success("Successfully tested application "+colors.Blue+"`%s`"+colors.Reset+" in "+colors.Yellow+"%.1f"+colors.Reset+" seconds", project.AppName, time.Since(tn).Seconds())

	return nil
}

func (g *GoBuilder) sumExec(project models.Project) error {
	for _, target := range g.targets {
		outputName := fmt.Sprintf("%s-%s-%s%s", project.AppName, target.GOOS, target.GOARCH, target.EXECSUFFIX)
		outputPath := filepath.Join(project.BuildDir, outputName)

		if _, err := os.Lstat(outputPath); os.IsNotExist(err) {
			colors.ErrLog("Application build in `%s` does not exist! Failed to generate SHA sums", outputPath)
			continue
		}

		contents, err := os.ReadFile(outputPath)
		if err != nil {
			colors.ErrLog("Cannot open build from `%s`: %v! Failed to generate SHA sums", outputPath, err)
			continue
		}

		for hasherLabel, hasher := range g.hashers {
			tn := time.Now()
			hasherLabelUpper := strings.ToUpper(hasherLabel)
			colors.Icon(colors.Yellow, "\u226b", "Generating %s Sum for app "+colors.Blue+"%s"+colors.Green+" (%s:%s)"+colors.Reset+" to %s", hasherLabelUpper, project.AppName+target.EXECSUFFIX+".sha265", target.GOOS, target.GOARCH, project.BuildDir)
			outputShaSum := outputPath + "." + hasherLabel

			if _, err := hasher.Write(contents); err != nil {
				colors.ErrLog("Cannot generate `%s` sum from `%s`: %v! Failed to generate SHA sums", hasherLabelUpper, outputPath, err)
				continue
			}
			buildSumHex := hex.EncodeToString(hasher.Sum(nil))
			hasher.Reset()
			if err := os.WriteFile(outputShaSum, []byte(buildSumHex), os.ModePerm); err != nil {
				colors.ErrLog("Cannot write %s sum file in `%s`: %v! Failed to generate SHA-256 sum", hasherLabelUpper, outputShaSum, err)
				continue
			}
			colors.Success("Generated "+colors.Green+"%s"+colors.Reset+" app "+colors.Blue+"`%s`"+colors.Reset+" for architecture "+colors.Green+"%s:%s"+colors.Reset+" in "+colors.Yellow+"%.1f"+colors.Reset+" seconds, output: %s", buildSumHex, project.AppName, target.GOOS, target.GOARCH, time.Since(tn).Seconds(), outputPath)
		}
	}
	return nil
}

func persistLog(buf bytes.Buffer, buf2 bytes.Buffer, dir string, name string) {
	outFileLog := filepath.Join(dir, fmt.Sprintf("build-%s.log", name))
	outErrLog := filepath.Join(dir, fmt.Sprintf("error-%s.log", name))

	for k, v := range map[string]*bytes.Buffer{
		outFileLog: &buf,
		outErrLog:  &buf2,
	} {
		if err := os.WriteFile(k, v.Bytes(), os.ModePerm); err != nil {
			fmt.Printf("Error writing build log: %v", err)
			continue
		}
		colors.ErrLog("Error building app "+colors.Red+"%s"+colors.Reset+"! Log stored in %s", name, k)
	}
}

func (g *GoBuilder) buildExec(project models.Project) error {
	for _, target := range g.targets {
		tn := time.Now()
		colors.Icon(colors.Yellow, "\u226b", "Building app "+colors.Blue+"%s"+colors.Green+" (%s:%s)"+colors.Reset+" to %s", project.AppName+target.EXECSUFFIX, target.GOOS, target.GOARCH, project.BuildDir)
		outputName := fmt.Sprintf("%s-%s-%s%s", project.AppName, target.GOOS, target.GOARCH, target.EXECSUFFIX)
		outputPath := filepath.Join(project.BuildDir, outputName)

		// Prepare the build command: go build -o outputPath project.AppMainSrcDir
		cmd := exec.Command(filepath.Join(g.goRootPath, "bin", "go"), "build", "-o", outputPath, project.AppMainSrcDir)
		cmd.Dir = project.RootDir
		env := append(g.defaultEnv, fmt.Sprintf("GOOS=%s", target.GOOS))
		env = append(env, fmt.Sprintf("GOARCH=%s", target.GOARCH))
		cmd.Env = env

		// Capture output
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		// Execute the command
		if err := cmd.Run(); err != nil {
			// If there's an error, return the captured stdout and stderr as part of the error
			return fmt.Errorf("error building for %s/%s: %v\n%s\n%s", target.GOOS, target.GOARCH, err, outBuf.String(), errBuf.String())
		}
		colors.Success("Successfully built app "+colors.Blue+"`%s`"+colors.Reset+" for architecture "+colors.Green+"%s:%s"+colors.Reset+" in "+colors.Yellow+"%.1f"+colors.Reset+" seconds, output: %s", project.AppName, target.GOOS, target.GOARCH, time.Since(tn).Seconds(), outputPath)
	}

	return nil
}

func loadEnvironmentFile() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		colors.ErrLog("Error getting user home directory. No `~/.go-env` configuraiton loaded: %v", err)
		return []string{}
	}

	if ifl, err := os.Lstat(filepath.Join(homeDir, ".go-env")); err != nil || ifl.IsDir() {
		return []string{}
	}

	contents, err := os.ReadFile(filepath.Join(homeDir, ".go-env"))
	if err != nil {
		return []string{}
	}

	colors.InfoLog("Additional environment configuration loaded from `~/.go-env` file")
	return strings.Split(string(contents), "\n")
}

func NewGoBuilder(toolchainPath string, conf models.SelectedConfig) *GoBuilder {
	targets := GoBuilderTargets{}

	for osName, osArch := range conf.Profile.OS {
		suffix := ""
		if osName == "windows" {
			suffix = ".exe"
		}
		for _, archTarget := range osArch {
			targets = append(targets, GoBuilderTarget{
				GOOS:       osName,
				GOARCH:     archTarget,
				EXECSUFFIX: suffix,
			})
		}

	}

	env := os.Environ()
	env = append(env, loadEnvironmentFile()...)

	stagesMap := map[string]bool{}
	for _, stage := range conf.Profile.Stages {
		stagesMap[stage] = true
	}

	return &GoBuilder{
		toolchainPath,
		filepath.Join(toolchainPath, "go"),
		filepath.Join(toolchainPath, "gopath"),
		targets,
		env,
		map[string]hash.Hash{
			"sha256": sha256.New(),
			"sha512": sha512.New(),
			"sha1":   sha1.New(),
		},
		stagesMap,
	}
}
