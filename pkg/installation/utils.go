package installation

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/fsouza/go-dockerclient"
	"github.com/kyma-project/cli/internal/minikube"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (i *Installation) applyResourceFile(filepath string) error {
	_, err := i.getKubectl().RunCmd("apply", "-f", filepath)
	return err
}

func (i *Installation) buildKymaInstaller(imageName string) error {
	dc, err := minikube.DockerClient(i.Options.Verbose, i.Options.LocalCluster.Profile)
	if err != nil {
		return err
	}

	var args []docker.BuildArg
	return dc.BuildImage(docker.BuildImageOptions{
		Name:         strings.TrimSpace(string(imageName)),
		Dockerfile:   filepath.Join("tools", "kyma-installer", "kyma.Dockerfile"),
		OutputStream: ioutil.Discard,
		ContextDir:   filepath.Join(i.Options.LocalSrcPath),
		BuildArgs:    args,
	})
}

func (i *Installation) checkIfResourcePresent(namespace, kind, name string) error {
	_, err := i.getKubectl().RunCmd("-n", namespace, "get", kind, name)
	return err
}

func (i *Installation) getInstallationStatus() (status string, desc string, err error) {
	status, err = i.getKubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.state}'")
	if err != nil {
		return
	}
	desc, err = i.getKubectl().RunCmd("get", "installation/kyma-installation", "-o", "jsonpath='{.status.description}'")
	return
}

func (i *Installation) printInstallationErrorLog() error {
	logs, err := i.getKubectl().RunCmd("get", "installation", "kyma-installation", "-o", "go-template", "--template={{- range .status.errorLog -}}{{printf \"%s:\\n %s [%s]\\n\" .component .log .occurrences}}{{- end}}")
	if err != nil {
		return err
	}
	fmt.Println(logs)
	return nil
}

func (i *Installation) getLatestAvailableMasterHash() (string, error) {
	ctx, timeoutF := context.WithTimeout(context.Background(), 2*time.Minute)
	defer timeoutF()
	maxCloningDepth := 5
	r, err := git.CloneContext(ctx, memory.NewStorage(), nil,
		&git.CloneOptions{
			Depth: maxCloningDepth,
			URL:   "https://github.com/kyma-project/kyma",
		})
	if err != nil {
		return "", errors.Wrap(err, "while cloning Kyma repository")
	}

	h, err := r.Head()
	if err != nil {
		return "", errors.Wrap(err, "while getting head of Kyma repository: %w")
	}

	iter, err := r.Log(&git.LogOptions{From: h.Hash()})
	if err != nil {
		return "", errors.Wrap(err, "while getting logs of Kyma repository: %w")
	}

	defer iter.Close()

	for i := 0; i < maxCloningDepth; i++ {
		c, err := iter.Next()
		if err != nil {
			return "", errors.Wrap(err, "while iterating commit of Kyma repository: %w")
		}
		if c == nil {
			return "", errors.New("while iterating commit of Kyma repository: commit is nil")
		}

		abbrevHash := c.Hash.String()[:8]
		resp, err := http.Head(fmt.Sprintf("https://storage.googleapis.com/kyma-development-artifacts/master-%s/kyma-installer-cluster.yaml", abbrevHash))
		if err != nil {
			return "", errors.Wrap(err, "while fetching example file from kyma-development-artifacts")
		}
		if err = resp.Body.Close(); err != nil {
			return "", errors.Wrap(err, "while closing body")
		}
		if resp.StatusCode == http.StatusOK {
			return abbrevHash, nil
		} else if resp.StatusCode != http.StatusNotFound {
			return "", fmt.Errorf("got unexpected status code when fetching example file from kyma-development artifacts, got: [%d] ", resp.StatusCode)
		}

	}

	return "", errors.New("not found latest available master hash")
}

func (i *Installation) setAdminPassword() error {
	if i.Options.Password == "" {
		return nil
	}
	encPass := base64.StdEncoding.EncodeToString([]byte(i.Options.Password))
	_, err := i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Patch("installation-config-overrides", types.JSONPatchType,
		[]byte(fmt.Sprintf("[{\"op\": \"replace\", \"path\": \"/data/global.adminPassword\", \"value\": \"%s\"}]", encPass)))
	if err != nil {
		if apiErrors.IsNotFound(err) {
			_, err = i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Create(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "installation-config-overrides",
					Labels: map[string]string{"installer": "overrides"},
				},
				Data: map[string]string{
					"global.adminPassword": encPass,
				},
			})
		}
		err = errors.Wrap(err, "Error setting admin password")
	}
	return err
}

func removeActionLabel(acc *[]map[string]interface{}) error {
	for _, config := range *acc {
		if kind, ok := config["kind"]; ok && kind == "Installation" {
			meta, ok := config["metadata"].(map[interface{}]interface{})
			if !ok {
				return errors.New("Installation contains no METADATA section")
			}

			labels, ok := meta["labels"].(map[interface{}]interface{})
			if !ok {
				return errors.New("Installation contains no LABELS section")
			}

			_, ok = labels["action"].(string)
			if !ok {
				return nil
			}

			delete(labels, "action")
		}
	}
	return nil
}

func buildDockerImageString(template string, version string) string {
	return fmt.Sprintf(template, version)
}

func downloadFile(path string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func getInstallerImage(resources *[]map[string]interface{}) (string, error) {
	for _, res := range *resources {
		if res["kind"] == "Deployment" {

			var deployment v1.Deployment
			err := mapstructure.Decode(res, &deployment)
			if err != nil {
				return "", err
			}

			if deployment.Spec.Template.Spec.Containers[0].Name == "kyma-installer-container" {
				return deployment.Spec.Template.Spec.Containers[0].Image, nil
			}
		}
	}
	return "", errors.New("'kyma-installer' deployment is missing")
}

func replaceInstallerImage(resources *[]map[string]interface{}, imageURL string) error {
	// Check if installer deployment has all the necessary fields and a container named kyma-installer-container.
	// If so, replace the image with the imageURL parameter.
	for _, config := range *resources {
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
	return errors.New("unable to find 'image' field for kyma installer 'Deployment'")
}

func isDockerImage(s string) bool {
	return len(strings.Split(s, "/")) > 1
}

func isSemVer(s string) bool {
	_, err := semver.NewVersion(s)
	return err == nil
}
