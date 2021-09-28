package istio

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultIstioChartPath = "/resources/istio-configuration/Chart.yaml"
const archSupport = "1.6"
const environmentVariable = "ISTIOCTL_PATH"

type Config struct {
	APIVersion    string   `yaml:"apiVersion"`
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	AppVersion    string   `yaml:"appVersion"`
	TillerVersion string   `yaml:"tillerVersion"`
	Description   string   `yaml:"description"`
	Keywords      []string `yaml:"keywords"`
	Sources       []string `yaml:"sources"`
	Engine        string   `yaml:"engine"`
	Home          string   `yaml:"home"`
	Icon          string   `yaml:"icon"`
}

type Installation struct {
	WorkspacePath  string
	IstioChartPath string
	environmentVar string
	istioVersion   string
	osext          string
	istioArch      string
	archSupport    string
	binPath string
}

func New(workspacePath string) Installation {
	return Installation{
		WorkspacePath:  workspacePath,
		IstioChartPath: defaultIstioChartPath,
		environmentVar: environmentVariable,
		archSupport:    archSupport,
	}
}

func (i *Installation) Install() error {
	// Get wanted Istio Version
	if err := i.getIstioVersion(); err != nil {
		return fmt.Errorf("error checking wanted istio version: %s", err)
	}
	// Check if Istioctl binary is already in workspace
	exist, err := i.checkIfExists()
	if err != nil {
		return err
	}
	if !exist {
		// Set OS Version
		i.setOS()
		// Set OS Architecture
		i.setArch()
		// Download Istioctl
		if err := i.downloadIstio(); err != nil {
			return fmt.Errorf("error downloading istio: %s", err)
		}
		// Extract tar.gz
		if err := i.extractIstio(); err != nil {
			return fmt.Errorf("error extracting istio.tar.gz: %s", err)
		}
	}
	// Export env variable
	if err := i.exportEnvVar(); err != nil {
		return fmt.Errorf("error exporting environment variable: %s", err)
	}
	return nil
}

func (i *Installation) getIstioVersion() error{
	var chart Config
	istioConfig, err := ioutil.ReadFile(filepath.Join(i.WorkspacePath, i.IstioChartPath))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(istioConfig, &chart)
	if err != nil {
		return err
	}
	i.istioVersion = chart.AppVersion
	i.binPath = path.Join(i.WorkspacePath, "istioctl", fmt.Sprintf("istio-%s", i.istioVersion), "bin", "istioctl")
	return nil
}

func (i *Installation) checkIfExists() (bool, error) {
	_, err := os.Stat(i.binPath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (i *Installation) setOS() {
	// Get OS Version
	i.osext = runtime.GOOS
	switch i.osext {
	case "windows":
		i.osext = "win"
	case "darwin":
		i.osext = "osx"
	default:
		i.osext = "linux"
	}
}

func (i *Installation) setArch() {
	i.istioArch = runtime.GOARCH
	if i.osext == "osx" && i.istioArch == "amd64"{
		i.istioArch = "arm64"
	}
}

func (i *Installation) downloadIstio() error {
	// Istioctl download links
	nonArchUrl := fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-%s.tar.gz", i.istioVersion, i.istioVersion, i.osext)
	archUrl := fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-%s-%s.tar.gz", i.istioVersion, i.istioVersion, i.osext, i.istioArch)

	if i.osext == "linux" {
		if strings.Split(i.archSupport, ".")[1] >= strings.Split(i.istioVersion, ".")[1] {
			err := downloadFile(path.Join(i.WorkspacePath, "istioctl"), "istio.tar.gz", archUrl)
			if err != nil {
				return err
			}
		} else {
			err := downloadFile(path.Join(i.WorkspacePath, "istioctl"), "istio.tar.gz", nonArchUrl)
			if err != nil {
				return err
			}
		}
	} else if i.osext == "osx" {
		err := downloadFile(path.Join(i.WorkspacePath, "istioctl"), "istio.tar.gz", nonArchUrl)
		if err != nil {
			return err
		}
	} else if i.osext == "win" {
		// TODO
	} else {
		// TODO
	}
	return nil
}

func (i *Installation) extractIstio() error {
	istioPath := path.Join(i.WorkspacePath, "istioctl", "istio.tar.gz")
	targetPath := path.Join(i.WorkspacePath, "istioctl", "istio.tar")
	if err := unGzip(istioPath, targetPath); err != nil {
		return err
	}
	istioPath = path.Join(i.WorkspacePath, "istioctl", "istio.tar")
	targetPath = path.Join(i.WorkspacePath, "istioctl")
	if err := unTar(istioPath, targetPath); err != nil {
		return err
	}
	return nil
}

func (i *Installation) exportEnvVar() error {
	if err := os.Setenv(i.environmentVar, i.binPath); err != nil {
		return err
	}
	return nil
}

func downloadFile(filepath string, filename string, url string) error {
	// Get data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create path and file
	err = os.MkdirAll(filepath, 0700)
	if err != nil {
		return err
	}
	out, err := os.Create(path.Join(filepath, filename))
	if err != nil {
		return err
	}
	defer out.Close()

	// Write body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func unGzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func unTar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		headerPath := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(headerPath, info.Mode()); err != nil {
				return err
			}
			continue
		}

		err = func() error {
			file, err := os.OpenFile(headerPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil{
			return err
		}
	}
	return nil
}
