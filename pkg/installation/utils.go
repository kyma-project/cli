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
	docker "github.com/fsouza/go-dockerclient"
	"github.com/kyma-project/cli/internal/minikube"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	v1 "k8s.io/api/apps/v1"
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
	logs, err := i.getKubectl().RunCmd("get", "installation", "kyma-installation", "-o", "go-template", "--template={{- range .status.errorLog -}}{{printf \"%s:\n %s [%s]\n\" .component .log .occurrences}}{{- end}}")
	if err != nil {
		return err
	}
	fmt.Println(logs)
	return nil
}

func (i *Installation) getMasterHash() (string, error) {
	ctx, timeoutF := context.WithTimeout(context.Background(), 1*time.Minute)
	defer timeoutF()
	r, err := git.CloneContext(ctx, memory.NewStorage(), nil,
		&git.CloneOptions{
			Depth: 1,
			URL:   "https://github.com/kyma-project/kyma",
		})
	if err != nil {
		return "", err
	}
	h, err := r.Head()
	if err != nil {
		return "", err
	}

	return h.Hash().String()[:8], nil
}

func (i *Installation) setAdminPassword() error {
	if i.Options.Password == "" {
		return nil
	}
	encPass := base64.StdEncoding.EncodeToString([]byte(i.Options.Password))
	_, err := i.k8s.Static().CoreV1().ConfigMaps("kyma-installer").Patch("installation-config-overrides", types.JSONPatchType,
		[]byte(fmt.Sprintf("[{\"op\": \"replace\", \"path\": \"/data/global.adminPassword\", \"value\": \"%s\"}]", encPass)))
	if err != nil {
		err = errors.Wrap(err, "Error setting admin password")
	}
	return err
}

func removeActionLabel(acc *[]map[string]interface{}) error {
	for _, config := range *acc {
		kind, ok := config["kind"]
		if !ok {
			continue
		}

		if kind != "Installation" {
			continue
		}

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
			mapstructure.Decode(res, &deployment)

			if deployment.Spec.Template.Spec.Containers[0].Name == "kyma-installer-container" {
				return deployment.Spec.Template.Spec.Containers[0].Image, nil
			}
		}
	}
	return "", errors.New("'kyma-installer' deployment is missing")
}

func replaceInstallerImage(resources *[]map[string]interface{}, imageURL string) error {
	for _, config := range *resources {
		kind, ok := config["kind"]
		if !ok {
			continue
		}

		if kind != "Deployment" {
			continue
		}

		spec, ok := config["spec"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		template, ok := spec["template"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		spec, ok = template["spec"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		if accName, ok := spec["serviceAccountName"]; !ok {
			continue
		} else {
			if accName != "kyma-installer" {
				continue
			}
		}

		containers, ok := spec["containers"].([]interface{})
		if !ok {
			continue
		}
		for _, c := range containers {
			container := c.(map[interface{}]interface{})
			cName, ok := container["name"]
			if !ok {
				continue
			}

			if cName != "kyma-installer-container" {
				continue
			}

			if _, ok := container["image"]; !ok {
				continue
			}
			container["image"] = imageURL
			return nil
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
