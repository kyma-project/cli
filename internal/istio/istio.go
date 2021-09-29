package istio

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/kyma-project/cli/internal/files"
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
const dirName = "istio"
const binName = "istioctl"
const winBinName = "istioctl.exe"
const downloadUrl = "https://github.com/istio/istio/releases/download/"
const tarGzName = "istio.tar.gz"
const tarName = "istio.tar"
const zipName = "istio.zip"

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

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
	Client HTTPClient
	kymaHome string
	environmentVar string
	istioVersion   string
	osExt          string
	istioArch      string
	archSupport    string
	binPath        string
	dirName        string
	binName        string
	winBinName     string
	downloadUrl    string
	tarGzName      string
	tarName        string
	zipName        string
}

func New(workspacePath string) (Installation, error) {
	kymaHome, err := files.KymaHome()
	if err != nil {
		return Installation{}, err
	}
	return Installation{
		WorkspacePath:  workspacePath,
		IstioChartPath: defaultIstioChartPath,
		Client: &http.Client{},
		kymaHome: kymaHome,
		environmentVar: environmentVariable,
		archSupport:    archSupport,
		dirName:        dirName,
		binName:        binName,
		winBinName:     winBinName,
		downloadUrl:    downloadUrl,
		tarGzName:      tarGzName,
		tarName:        tarName,
		zipName:        zipName,
	}, nil
}

func (i *Installation) Install() error {
	// Set OS Version
	i.setOS()
	// Set OS Architecture
	i.setArch()
	// Get wanted Istio Version
	if err := i.getIstioVersion(); err != nil {
		return fmt.Errorf("error checking wanted istio version: %s", err)
	}
	// Check if Istioctl binary is already in kymaHome
	exist, err := i.checkIfExists()
	if err != nil {
		return err
	}
	if !exist {
		// Download Istioctl
		if err := i.downloadIstio(); err != nil {
			return fmt.Errorf("error downloading istio: %s", err)
		}
		// Extract tar.gz or zip
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

func (i *Installation) getIstioVersion() error {
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
	fmt.Printf("IstioVersion: %s\n", i.istioVersion)
	if i.istioVersion == "" {
		return errors.New("istio version is empty")
	}
	if i.osExt == "win" {
		// TODO Windows: Test if this is correct path
		i.binPath = path.Join(i.kymaHome, i.dirName, fmt.Sprintf("istio-%s", i.istioVersion), i.winBinName)
	} else {
		i.binPath = path.Join(i.kymaHome, i.dirName, fmt.Sprintf("istio-%s", i.istioVersion), "bin", i.binName)
	}
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
	i.osExt = runtime.GOOS
	switch i.osExt {
	case "windows":
		i.osExt = "win"
	case "darwin":
		i.osExt = "osx"
	default:
		i.osExt = "linux"
	}
}

func (i *Installation) setArch() {
	i.istioArch = runtime.GOARCH
	if i.osExt == "osx" && i.istioArch == "amd64" {
		i.istioArch = "arm64"
	}
}

func (i *Installation) downloadIstio() error {
	// Istioctl download links
	nonArchUrl := fmt.Sprintf("%s%s/istio-%s-%s.tar.gz", downloadUrl, i.istioVersion, i.istioVersion, i.osExt)
	archUrl := fmt.Sprintf("%s%s/istio-%s-%s-%s.tar.gz", downloadUrl, i.istioVersion, i.istioVersion, i.osExt, i.istioArch)

	if i.osExt == "linux" {
		if strings.Split(i.archSupport, ".")[1] >= strings.Split(i.istioVersion, ".")[1] {
			err := i.downloadFile(path.Join(i.kymaHome, dirName), tarGzName, archUrl)
			if err != nil {
				return err
			}
		} else {
			err := i.downloadFile(path.Join(i.kymaHome, dirName), tarGzName, nonArchUrl)
			if err != nil {
				return err
			}
		}
	} else if i.osExt == "osx" {
		err := i.downloadFile(path.Join(i.kymaHome, dirName), tarGzName, nonArchUrl)
		if err != nil {
			return err
		}
	} else if i.osExt == "win" {
		err := i.downloadFile(path.Join(i.kymaHome, dirName), zipName, nonArchUrl)
		if err != nil {
			return err
		}
	} else {
		return errors.New("unsupported operating system")
	}
	return nil
}

func (i *Installation) extractIstio() error {
	if i.osExt == "linux" || i.osExt == "osx" {
		istioPath := path.Join(i.kymaHome, i.dirName, i.tarGzName)
		targetPath := path.Join(i.kymaHome, i.dirName, i.tarName)
		if err := unGzip(istioPath, targetPath, true); err != nil {
			return err
		}
		istioPath = path.Join(i.kymaHome, i.dirName, i.tarName)
		targetPath = path.Join(i.kymaHome, i.dirName)
		if err := unTar(istioPath, targetPath, true); err != nil {
			return err
		}
	} else {
		istioPath := path.Join(i.kymaHome, i.dirName, i.zipName)
		targetPath := path.Join(i.kymaHome, i.dirName)
		if err := unZip(istioPath, targetPath, true); err != nil {
			return err
		}
	}
	return nil
}

func (i *Installation) exportEnvVar() error {
	if err := os.Setenv(i.environmentVar, i.binPath); err != nil {
		return err
	}
	return nil
}

func (i *Installation) downloadFile(filepath string, filename string, url string) error {
	// Get data
	resp, err := i.Client.Get(url)
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

func unGzip(source, target string, deleteSource bool) error {
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
	if err != nil {
		return err
	}
	if deleteSource {
		if err := os.Remove(source); err != nil {
			return err
		}
	}
	return err
}

func unTar(source, target string, deleteSource bool) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
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
		if err != nil {
			return err
		}
	}
	if deleteSource {
		if err := os.Remove(source); err != nil {
			return err
		}
	}
	return nil
}

func unZip(source, target string, deleteSource bool) error {
	// TODO Windows: Test + Unit Tests
	archive, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(target, f.Name)
		fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(target)+string(os.PathSeparator)) {

			return errors.New("invalid file path")
		}
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	if deleteSource {
		if err := os.Remove(source); err != nil {
			return err
		}
	}
	return nil
}
