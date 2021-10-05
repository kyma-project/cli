package installation

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/kyma-incubator/hydroform/install/config"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/kyma-incubator/hydroform/install/scheme"
	"github.com/kyma-project/cli/internal/resolve"
	"github.com/kyma-project/cli/pkg/step"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
)

func getLatestAvailableMainHash(currentStep step.Step, fallbackLevel int, nonInteractive bool) (string, error) {
	ctx, timeoutF := context.WithTimeout(context.Background(), 2*time.Minute)
	defer timeoutF()
	maxCloningDepth := fallbackLevel + 1
	r, err := git.CloneContext(ctx, memory.NewStorage(), nil,
		&git.CloneOptions{
			Depth:      maxCloningDepth,
			URL:        "https://github.com/kyma-project/kyma",
			NoCheckout: true,
		})
	if err != nil {
		return "", errors.Wrap(err, "while cloning Kyma repository")
	}

	h, err := r.Head()
	if err != nil {
		return "", errors.Wrap(err, "while getting head of Kyma repository: %w")
	}

	headHash := h.Hash().String()[:8]
	if fallbackLevel == 0 {
		return headHash, nil
	}
	artifactsAvailable, err := checkArtifactsAvailability(headHash)
	if err != nil {
		return "", errors.Wrap(err, "while checking availability of artifacts")
	}
	if artifactsAvailable {
		return headHash, nil
	} else if !nonInteractive {
		promptMsg := fmt.Sprintf("Artifacts for main-%s are not available. Would you like to use artifacts from previous commits?", headHash)
		if proceed := currentStep.PromptYesNo(promptMsg); !proceed {
			return "", errors.Errorf("aborting")
		}
	}

	iter, err := r.Log(&git.LogOptions{From: h.Hash()})
	if err != nil {
		return "", errors.Wrap(err, "while getting logs of Kyma repository: %w")
	}

	defer iter.Close()
	for k := 0; k < maxCloningDepth; k++ {
		c, err := iter.Next()
		if err != nil {
			return "", errors.Wrap(err, "while iterating commit of Kyma repository: %w")
		}
		if c == nil {
			return "", errors.New("while iterating commit of Kyma repository: commit is nil")
		}

		commitHash := c.Hash.String()[:8]
		artifactsAvailable, err := checkArtifactsAvailability(commitHash)
		if err != nil {
			return "", errors.Wrap(err, "while checking availability of artifacts")
		}
		if artifactsAvailable {
			return commitHash, nil
		}
		currentStep.LogInfof("Skipping version: [%s]: artifacts not yet available", commitHash)

	}

	return "", errors.New("unable to find a commit with available artifacts")
}

func checkArtifactsAvailability(abbrevHash string) (bool, error) {
	resp, err := http.Head(fmt.Sprintf(releaseResourcePattern, developmentBucket, "main-"+abbrevHash, "kyma-installer-cluster.yaml"))
	if err != nil {
		return false, errors.Wrap(err, "while fetching example file from kyma-development-artifacts")
	}
	if err = resp.Body.Close(); err != nil {
		return false, errors.Wrap(err, "while closing body")
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode != http.StatusNotFound {
		return false, fmt.Errorf("got unexpected status code when fetching example file from kyma-development artifacts, got: [%d] ", resp.StatusCode)
	}
	return false, nil
}

func (i *Installation) loadInstallationFiles() (map[string]*File, error) {
	var installationFiles map[string]*File
	if i.Options.fromLocalSources {
		if i.Options.IsLocal {
			installationFiles =
				map[string]*File{
					installerFile:       {Path: "installer.yaml"},
					installerCRFile:     {Path: "installer-cr.yaml.tpl"},
					installerConfigFile: {Path: "installer-config-local.yaml.tpl"},
				}
		} else {
			installationFiles =
				map[string]*File{
					installerFile:   {Path: "installer.yaml"},
					installerCRFile: {Path: "installer-cr-cluster.yaml.tpl"},
				}
		}
	} else {
		if i.Options.IsLocal {
			installationFiles =
				map[string]*File{
					installerFile:       {Path: "kyma-installer-cluster.yaml"},
					installerCRFile:     {Path: "kyma-installer-cr-local.yaml"},
					installerConfigFile: {Path: "kyma-config-local.yaml"},
				}
		} else {
			installationFiles =
				map[string]*File{
					installerFile:   {Path: "kyma-installer-cluster.yaml"},
					installerCRFile: {Path: "kyma-installer-cr-cluster.yaml"},
				}
		}
	}

	for _, file := range installationFiles {
		resources := make([]map[string]interface{}, 0)
		var reader io.ReadCloser
		var err error
		if i.Options.fromLocalSources {
			path := filepath.Join(i.Options.LocalSrcPath, "installation",
				"resources", file.Path)
			reader, err = os.Open(path)
		} else {
			reader, err = resolve.RemoteReader(i.releaseFile(file.Path))
		}

		if err != nil {
			return nil, err
		}

		dec := yaml.NewDecoder(reader)
		for {
			m := make(map[string]interface{})
			err := dec.Decode(m)
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			resources = append(resources, m)
		}

		reader.Close()
		file.Content = resources
	}

	return installationFiles, nil
}

func loadStringContent(installationFiles map[string]*File) (map[string]*File, error) {
	for _, file := range installationFiles {
		if file.Content != nil {
			buf := &bytes.Buffer{}
			enc := yaml.NewEncoder(buf)
			for _, y := range file.Content {
				err := enc.Encode(y)
				if err != nil {
					return installationFiles, err
				}
			}

			err := enc.Close()
			if err != nil {
				return installationFiles, err
			}

			file.StringContent = buf.String()
		}
	}

	return installationFiles, nil
}

func (i *Installation) loadConfigurations(files map[string]*File) (installationSDK.Configuration, error) {
	var configuration installationSDK.Configuration
	var configFileContent string
	for _, file := range i.Options.OverrideConfigs {
		oFile, err := os.Open(file)
		if err != nil {
			return configuration, fmt.Errorf("error: unable to open file: %s", err.Error())
		}

		var rawData bytes.Buffer
		if _, err = io.Copy(&rawData, oFile); err != nil {
			fmt.Printf("unable to read data from file: %s.\n", file)
		}

		configFileContent = rawData.String() + "\n---\n" + configFileContent
	}

	if i.Options.IsLocal {
		//Merge with local config file
		configFileContent = files[installerConfigFile].StringContent + "\n---\n" + configFileContent
	}

	if configFileContent != "" {
		decoder, err := scheme.DefaultDecoder()
		if err != nil {
			return configuration, fmt.Errorf("error: failed to create default decoder: %s", err.Error())
		}

		configuration, err = config.YAMLToConfiguration(decoder, configFileContent)
		if err != nil {
			return configuration, fmt.Errorf("error: failed to parse configurations: %s", err.Error())
		}
	}

	if i.Options.IsLocal {
		configuration.Configuration.Set("global.minikubeIP", i.Options.LocalCluster.IP, false)
	}
	if i.Options.Password != "" {
		configuration.Configuration.Set("global.adminPassword", base64.StdEncoding.EncodeToString([]byte(i.Options.Password)), false)
	}
	if i.Options.Domain != "" && i.Options.Domain != defaultDomain {
		configuration.Configuration.Set("global.domainName", i.Options.Domain, false)
		configuration.Configuration.Set("global.tlsCrt", i.Options.TLSCert, false)
		configuration.Configuration.Set("global.tlsKey", i.Options.TLSKey, false)
	}

	return configuration, nil
}

func LoadComponentsConfig(cfgFile string) ([]v1alpha1.KymaComponent, error) {
	if cfgFile != "" {
		data, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return nil, err
		}

		var installationCR v1alpha1.Installation
		err = yaml.Unmarshal(data, &installationCR)
		if err != nil || len(installationCR.Spec.Components) < 1 {
			var config ComponentsConfig
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				return nil, err
			}

			return config.Components, nil
		}

		return installationCR.Spec.Components, nil
	}

	return []v1alpha1.KymaComponent{}, nil
}

func insertProfile(installationCRFile *File, profile string) error {
	for _, config := range installationCRFile.Content {
		if kind, ok := config["kind"]; ok && kind == "Installation" {
			if spec, ok := config["spec"].(map[interface{}]interface{}); ok {
				spec["profile"] = profile
				return nil
			}
		}
	}
	return errors.New("unable to set 'profile' field for Kyma Installation CR")
}

func getInstallerImage(installerFile *File) (string, error) {
	for _, config := range installerFile.Content {
		if kind, ok := config["kind"]; ok && kind == "Deployment" {
			if spec, ok := config["spec"].(map[interface{}]interface{}); ok {
				if template, ok := spec["template"].(map[interface{}]interface{}); ok {
					if spec, ok = template["spec"].(map[interface{}]interface{}); ok {
						if containers, ok := spec["containers"].([]interface{}); ok {
							for _, c := range containers {
								container := c.(map[interface{}]interface{})
								if cName, ok := container["name"]; ok && cName == "kyma-installer-container" {
									if _, ok := container["image"]; ok {
										return container["image"].(string), nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return "", errors.New("'kyma-installer' deployment is missing")
}

func replaceInstallerImage(installerFile *File, imageURL string) error {
	// Check if installer deployment has all the necessary fields and a container named kyma-installer-container.
	// If so, replace the image with the imageURL parameter.
	for _, config := range installerFile.Content {
		if kind, ok := config["kind"]; ok && kind == "Deployment" {
			if spec, ok := config["spec"].(map[interface{}]interface{}); ok {
				if template, ok := spec["template"].(map[interface{}]interface{}); ok {
					if spec, ok = template["spec"].(map[interface{}]interface{}); ok {
						if containers, ok := spec["containers"].([]interface{}); ok {
							for _, c := range containers {
								container := c.(map[interface{}]interface{})
								if cName, ok := container["name"]; ok && cName == "kyma-installer-container" {
									if _, ok := container["image"]; ok {
										container["image"] = imageURL
										return nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return errors.New("unable to find 'image' field for Kyma Installer 'Deployment'")
}

func isDockerImage(s string) bool {
	return len(strings.Split(s, "/")) > 1
}

func isSemVer(s string) bool {
	_, err := semver.Parse(s)
	return err == nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil && len(s) > 7
}
