package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/communitymodules"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

func ApplySpecifiedModules(ctx context.Context, client rootlessdynamic.Interface, modules, crs []string) clierror.Error {
	available, err := communitymodules.GetAvailableModules()
	if err != nil {
		return err
	}

	customConfig, err := readCustomConfig(crs)
	if err != nil {
		return err
	}

	for _, rec := range available {
		versionedName := containsModule(rec.Name, modules) //TODO move splitting to earlier
		if versionedName == nil {
			continue
		}

		wantedVersion := verifyVersion(versionedName, rec)
		fmt.Printf("Applying %s module manifest\n", rec.Name)
		err = applyGivenObjects(ctx, client, wantedVersion.DeploymentYaml)
		if err != nil {
			return err
		}

		if applyGivenCustomCR(ctx, client, rec, customConfig) {
			fmt.Println("Applying custom CR")
			continue
		}

		fmt.Println("Applying CR")
		err = applyGivenObjects(ctx, client, wantedVersion.CrYaml)
		if err != nil {
			return err
		}
	}
	return nil
}

func readCustomConfig(cr []string) ([]unstructured.Unstructured, clierror.Error) {
	if len(cr) == 0 {
		return nil, nil
	}
	var objects []unstructured.Unstructured
	for _, rec := range cr {
		yaml, err := os.ReadFile(rec)
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New("failed to read custom file"))
		}
		currentObjects, err := decodeYaml(bytes.NewReader(yaml))
		if err != nil {
			return nil, clierror.Wrap(err, clierror.New("failed to decode custom YAML"))
		}
		objects = append(objects, currentObjects...)
	}
	return objects, nil
}

func containsModule(have string, want []string) []string {
	for _, rec := range want {
		name := strings.Split(rec, ":")
		if name[0] == have {
			return name
		}
	}
	return nil
}

func verifyVersion(name []string, rec communitymodules.Module) communitymodules.Version {
	if len(name) != 1 {
		for _, version := range rec.Versions {
			if version.Version == name[1] {
				fmt.Printf("Version %s found for %s\n", version.Version, rec.Name)
				return version
			}
		}
	}

	fmt.Printf("Using latest version for %s\n", rec.Name)
	return communitymodules.GetLatestVersion(rec.Versions)
}

// applyGivenCustomCR applies custom CR if it exists
func applyGivenCustomCR(ctx context.Context, client rootlessdynamic.Interface, rec communitymodules.Module, config []unstructured.Unstructured) bool {
	for _, obj := range config {
		if strings.EqualFold(obj.GetKind(), strings.ToLower(rec.Name)) {
			client.Apply(ctx, &obj)
			return true
		}
	}
	return false

}

func applyGivenObjects(ctx context.Context, client rootlessdynamic.Interface, url string) clierror.Error {
	givenYaml, err := http.Get(url)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get YAML from URL"))
	}
	defer givenYaml.Body.Close()

	yamlContent, err := io.ReadAll(givenYaml.Body)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to read YAML"))
	}

	objects, err := decodeYaml(bytes.NewReader(yamlContent))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to decode YAML"))
	}

	cliErr := client.ApplyMany(ctx, objects)
	if cliErr != nil {
		return clierror.WrapE(cliErr, clierror.New("failed to apply module resources"))

	}
	return nil
}

func decodeYaml(r io.Reader) ([]unstructured.Unstructured, error) {
	results := make([]unstructured.Unstructured, 0)
	decoder := yaml.NewDecoder(r)

	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		u := unstructured.Unstructured{Object: obj}
		if u.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" {
			results = append([]unstructured.Unstructured{u}, results...)
			continue
		}
		results = append(results, u)
	}

	return results, nil
}
