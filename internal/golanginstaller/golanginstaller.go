package golanginstaller

import (
	"archive/tar"
	"archive/zip"
	"autobuild-go/internal/colors"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GoInstaller defines the structure for installing Go
type GoInstaller struct {
	projectPath  string
	toolchainDir string
}

// New creates a new instance of GoInstaller
func New(projectPath string) *GoInstaller {
	toolchainDir := filepath.Join(projectPath, ".toolchain")
	return &GoInstaller{
		projectPath:  projectPath,
		toolchainDir: toolchainDir,
	}
}

// EnsureGo checks if Go is installed, if not it installs the latest version
func (g *GoInstaller) EnsureGo() error {
	if g.isGoInstalled() {
		colors.Success("Go is already installed")
		return nil
	}

	colors.Icon(colors.Yellow, "\u226b", "Go is not installed. Installing latest version...")

	latestVersion, err := GetLatestGoVersion()
	if err != nil {
		return fmt.Errorf("error fetching latest Go version: %v", err)
	}

	if err := g.downloadAndInstallGo(latestVersion); err != nil {
		return fmt.Errorf("error downloading and installing Go: %v", err)
	}

	colors.Success("Go installed successfully!")
	return nil
}

// isGoInstalled checks if Go is already installed
func (g *GoInstaller) isGoInstalled() bool {
	_, err := os.Stat(filepath.Join(g.toolchainDir, "go"))
	return !os.IsNotExist(err)
}

// downloadAndInstallGo downloads and installs the latest Go version
func (g *GoInstaller) downloadAndInstallGo(version string) error {
	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	var archiveExt string
	var goFilename string

	if runtime.GOOS == "windows" {
		archiveExt = "zip"
		goFilename = fmt.Sprintf("go%s.%s.%s", version, osArch, archiveExt)
	} else {
		archiveExt = "tar.gz"
		goFilename = fmt.Sprintf("go%s.%s.%s", version, osArch, archiveExt)
	}

	downloadURL := fmt.Sprintf("https://golang.org/dl/%s", goFilename)

	// Download the archive
	archiveFilePath := filepath.Join(os.TempDir(), goFilename)
	if err := downloadFile(downloadURL, archiveFilePath); err != nil {
		return fmt.Errorf("error downloading Go archive: %v", err)
	}
	defer os.Remove(archiveFilePath) // Cleanup

	// Extract the file
	if runtime.GOOS == "windows" {
		if err := unzip(archiveFilePath, g.toolchainDir); err != nil {
			return fmt.Errorf("error unzipping Go archive: %v", err)
		}
	} else {
		if err := untarGz(archiveFilePath, g.toolchainDir); err != nil {
			return fmt.Errorf("error untarring Go archive: %v", err)
		}
	}

	return nil
}

// downloadFile downloads a file from a given URL to a local path
func downloadFile(url string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading file: %v", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// unzip extracts a zip file to the destination directory (for Windows)
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// untarGz extracts a tar.gz file to the destination directory (for Linux/macOS)
func untarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			return fmt.Errorf("unknown file type: %v in tar.gz", header.Typeflag)
		}
	}
	return nil
}
