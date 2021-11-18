package istioctl

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/kyma-project/cli/internal/files"
	"github.com/pkg/errors"
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
const downloadURL = "https://github.com/istio/istio/releases/download/"
const tarGzName = "istio.tar.gz"
const tarName = "istio.tar"
const zipName = "istio.zip"

type operatingSystem struct {
	name string
	ext  string
}

var windows = operatingSystem{"windows", "win"}
var darwin = operatingSystem{"darwin", "osx"}
var linux = operatingSystem{"linux", "linux"}

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
	Client         HTTPClient
	kymaHome       string
	environmentVar string
	istioVersion   string
	osExt          string
	istioArch      string
	archSupport    string
	binPath        string
	dirName        string
	binName        string
	winBinName     string
	downloadURL    string
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
		Client:         &http.Client{},
		kymaHome:       kymaHome,
		environmentVar: environmentVariable,
		archSupport:    archSupport,
		dirName:        dirName,
		binName:        binName,
		winBinName:     winBinName,
		downloadURL:    downloadURL,
		tarGzName:      tarGzName,
		tarName:        tarName,
		zipName:        zipName,
	}, nil
}

func (i *Installation) Install() error {
	if err := i.setOS(); err != nil {
		return err
	}
	i.setArch()
	if err := i.getIstioVersion(); err != nil {
		return errors.Wrap(err, "failed to get istio version")
	}
	exist, err := i.checkIfBinaryExists()
	if err != nil {
		return err
	}
	if !exist {
		if err := i.downloadIstio(); err != nil {
			return errors.Errorf("error downloading istioctl: %s", err)
		}
		if err := i.extractIstio(); err != nil {
			return errors.Errorf("error extracting istioctl: %s", err)
		}
	}
	if err := i.exportEnvVar(); err != nil {
		return errors.Errorf("error exporting environment variable: %s", err)
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
	if i.istioVersion == "" {
		return errors.New("istio version is empty")
	}
	if i.osExt == windows.ext {
		i.binPath = filepath.Join(i.kymaHome, i.dirName, fmt.Sprintf("istio-%s", i.istioVersion), "bin", i.winBinName)
	} else {
		i.binPath = filepath.Join(i.kymaHome, i.dirName, fmt.Sprintf("istio-%s", i.istioVersion), "bin", i.binName)
	}
	return nil
}

func (i *Installation) checkIfBinaryExists() (bool, error) {
	_, err := os.Stat(i.binPath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (i *Installation) setOS() error {
	i.osExt = runtime.GOOS
	switch i.osExt {
	case windows.name:
		i.osExt = windows.ext
	case darwin.name:
		i.osExt = darwin.ext
	case linux.name:
		i.osExt = linux.ext
	default:
		return errors.Errorf("unknown OS: %s", i.osExt)
	}
	return nil
}

func (i *Installation) setArch() {
	i.istioArch = runtime.GOARCH
	if i.osExt == darwin.ext && i.istioArch == "amd64" {
		i.istioArch = "arm64"
	}
}

func (i *Installation) downloadIstio() error {
	nonArchURL := ""
	archURL := ""
	// Istioctl download links
	if i.osExt == darwin.ext || i.osExt == linux.ext {
		nonArchURL = fmt.Sprintf("%s%s/istio-%s-%s.tar.gz", i.downloadURL, i.istioVersion, i.istioVersion, i.osExt)
		archURL = fmt.Sprintf("%s%s/istio-%s-%s-%s.tar.gz", i.downloadURL, i.istioVersion, i.istioVersion, i.osExt, i.istioArch)
	} else {
		nonArchURL = fmt.Sprintf("%s%s/istio-%s-%s.zip", i.downloadURL, i.istioVersion, i.istioVersion, i.osExt)
	}

	switch i.osExt {
	case linux.ext:
		if strings.Split(i.archSupport, ".")[1] >= strings.Split(i.istioVersion, ".")[1] {
			err := i.downloadFile(path.Join(i.kymaHome, dirName), tarGzName, archURL)
			if err != nil {
				return err
			}
		} else {
			err := i.downloadFile(path.Join(i.kymaHome, dirName), tarGzName, nonArchURL)
			if err != nil {
				return err
			}
		}
	case darwin.ext:
		err := i.downloadFile(path.Join(i.kymaHome, dirName), tarGzName, nonArchURL)
		if err != nil {
			return err
		}
	case windows.ext:
		err := i.downloadFile(path.Join(i.kymaHome, dirName), zipName, nonArchURL)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported operating system")
	}
	return nil
}

func (i *Installation) extractIstio() error {
	if i.osExt == linux.ext || i.osExt == darwin.ext {
		istioPath := filepath.Join(i.kymaHome, i.dirName, i.tarGzName)
		targetPath := filepath.Join(i.kymaHome, i.dirName, i.tarName)
		if err := unGzip(istioPath, targetPath, true); err != nil {
			return err
		}
		istioPath = filepath.Join(i.kymaHome, i.dirName, i.tarName)
		targetPath = filepath.Join(i.kymaHome, i.dirName)
		if err := unTar(istioPath, targetPath, true); err != nil {
			return err
		}
	} else {
		istioPath := filepath.Join(i.kymaHome, i.dirName, i.zipName)
		targetPath := filepath.Join(i.kymaHome, i.dirName)
		if err := unZip(istioPath, targetPath, true); err != nil {
			return err
		}
	}
	return nil
}

func (i *Installation) exportEnvVar() error {
	if i.environmentVar == "" || i.binPath == "" {
		return errors.New("envVar or binPath empty")
	}
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
	err = copyInChunks(out, resp.Body)
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

	err = copyInChunks(writer, archive)
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

		headerPath := fmt.Sprintf("%s/%s", target, header.Name)
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
			err = copyInChunks(file, tarReader)
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
	source = filepath.Clean(source)
	zipReader, err := initReader(source)
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		filePath, err := sanitizeExtractPath(target, f.Name)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filepath.Clean(filePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}
		defer fileInArchive.Close()

		if err := copyInChunks(dstFile, fileInArchive); err != nil {
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

func copyInChunks(dstFile *os.File, srcFile io.Reader) error {
	for {
		_, err := io.CopyN(dstFile, srcFile, 1024)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
	}
	return nil
}

func sanitizeExtractPath(destination, filePath string) (string, error) {
	destpath := filepath.Join(destination, filePath)
	if strings.Contains(destpath, "..") {
		return "", errors.Errorf("illegal destination path: %s", destpath)
	}
	if !strings.HasPrefix(destpath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return "", errors.Errorf("illegal destination path: %s", destpath)
	}
	return destpath, nil
}

func initReader(source string) (*zip.Reader, error) {
	ioReader, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer ioReader.Close()
	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, ioReader)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(buff.Bytes())
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, err
	}
	return zipReader, nil
}
