package builder

import (
	"autobuild-go/internal/colors"
	"autobuild-go/internal/models"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type GoBuilderTargets []struct {
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
			if err := g.testExec(project); err != nil {
				colors.ErrLog("Error: %v", err)
				return
			}
			if err := g.buildExec(project); err != nil {
				colors.ErrLog("Error building app %s: %v", project.AppName, err)
				return
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

func NewGoBuilder(toolchainPath string) *GoBuilder {
	targets := GoBuilderTargets{
		{"windows", "amd64", ".exe"},
		{"windows", "arm64", ".exe"},
		{"linux", "amd64", ""},
		{"linux", "arm64", ""},
		{"darwin", "amd64", ""},
		{"darwin", "arm64", ""},
	}

	env := os.Environ()
	env = append(env, loadEnvironmentFile()...)

	return &GoBuilder{
		toolchainPath: toolchainPath,
		goRootPath:    filepath.Join(toolchainPath, "go"),
		goPathPath:    filepath.Join(toolchainPath, "gopath"),
		targets:       targets,
		defaultEnv:    env,
	}
}
