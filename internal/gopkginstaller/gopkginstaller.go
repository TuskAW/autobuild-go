package gopkginstaller

import (
	"autobuild-go/internal/colors"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type pkgInstaller struct {
	packages   map[string]string
	goexecPath string
	gorootPath string
	gopathPath string
}

func (p *pkgInstaller) Install() error {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	for k, v := range p.packages {
		fmt.Printf("\t%s>>%sChecking package %s%s%s... ", colors.Green, colors.Reset, colors.Blue, k, colors.Reset)

		if _, err := os.Lstat(filepath.Join(p.gopathPath, "bin", fmt.Sprintf("%s%s", k, suffix))); err != nil {
			if p.installPkg(v) != nil {
				fmt.Println(colors.Red, "[fail]", colors.Reset)
				continue
			}
			fmt.Println(colors.Green, "[ok]", colors.Reset)
			continue
		}
		fmt.Println(colors.Green, "[ok]", colors.Reset)
	}
	return nil
}

func (p *pkgInstaller) installPkg(v string) error {
	envVariables := os.Environ()
	envVariables = append(envVariables, "GOPATH="+p.gopathPath, "GOROOT="+p.gorootPath)
	envVariables = append(envVariables, "PATH="+fmt.Sprintf("%s%s%s", filepath.Join(p.gopathPath, "bin"), string(os.PathListSeparator), os.Getenv("PATH")))

	execCmd := exec.Command(p.goexecPath, "install", v)
	execCmd.Env = envVariables

	execBuff := new(bytes.Buffer)
	execCmd.Stdout = execBuff
	execCmd.Stderr = execBuff

	if err := execCmd.Run(); err != nil {
		return err
	}
	return nil
}

func New(toolchainDir string, packages map[string]string) *pkgInstaller {
	goExec := "go"
	if runtime.GOOS == "windows" {
		goExec = "go.exe"
	}
	return &pkgInstaller{
		packages:   packages,
		goexecPath: filepath.Join(toolchainDir, "go", "bin", goExec),
		gorootPath: filepath.Join(toolchainDir, "go"),
		gopathPath: filepath.Join(toolchainDir, "gopath"),
	}
}
