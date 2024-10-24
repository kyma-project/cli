package extension

import (
	"context"
	"errors"
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListFromCluster(ctx context.Context, client kubernetes.Interface) ([]Extension, error) {
	labelSelector := fmt.Sprintf("%s==%s", extensionLabelKey, extensionResourceLabelValue)
	cms, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to load ConfigMaps from cluster with label %s", labelSelector)
	}

	extensions := make([]Extension, len(cms.Items))
	var parseErr error
	for i, cm := range cms.Items {
		extension, err := parseResourceExtension(cm.Data)
		if err != nil {
			// if the parse failed add an error to the errors list to take another extension
			// corrupted extension should not stop parsing the rest of the extensions
			parseErr = errors.Join(
				parseErr,
				pkgerrors.Wrapf(err, "failed to parse configmap '%s/%s'", cm.GetNamespace(), cm.GetName()),
			)
			continue
		}

		extensions[i] = *extension
	}

	return extensions, parseErr
}

func parseResourceExtension(cmData map[string]string) (*Extension, error) {
	rootCommand, err := parseRequiredField[RootCommand](cmData, extensionRootCommandKey)
	if err != nil {
		return nil, err
	}

	resourceInfo, err := parseOptionalField[ResourceInfo](cmData, extensionResourceInfoKey)
	if err != nil {
		return nil, err
	}

	genericCommands, err := parseOptionalField[GenericCommands](cmData, extensionGenericCommandsKey)
	if err != nil {
		return nil, err
	}

	return &Extension{
		RootCommand:     *rootCommand,
		ManagedResource: resourceInfo,
		GenericCommands: genericCommands,
	}, nil
}

func parseRequiredField[T any](cmData map[string]string, cmKey string) (*T, error) {
	dataBytes, ok := cmData[cmKey]
	if !ok {
		return nil, fmt.Errorf("missing .data.%s field", cmKey)
	}

	var data T
	err := yaml.Unmarshal([]byte(dataBytes), &data)
	return &data, err
}

func parseOptionalField[T any](cmData map[string]string, cmKey string) (*T, error) {
	dataBytes, ok := cmData[cmKey]
	if !ok {
		// skip because field is not required
		return nil, nil
	}

	var data T
	err := yaml.Unmarshal([]byte(dataBytes), &data)
	return &data, err
}
