package deploy

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/kyma-project/cli/internal/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const kymaCRTemplate = `---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
---
apiVersion: operator.kyma-project.io/v1alpha1
kind: Kyma
metadata:
  annotations:
    cli.kyma-project.io/source: deploy
  labels:
    operator.kyma-project.io/managed-by: lifecycle-manager
  name: default-kyma
  namespace: {{ .Namespace }}
spec:
  channel: {{ .Channel }}
  modules: []
  sync:
    enabled: {{ .Sync }}
`

var KymaGVR = schema.GroupVersionResource{
	Group:    "operator.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "kymas",
}

// Kyma deploys the Kyma CR. If no kymaCRPath is provided, it deploys the default CR.
func Kyma(k8s kube.KymaKube, namespace, channel, kymaCRpath string, dryRun bool) error {
	// TODO delete deploy.go when the old reconciler is gone.
	yamlBytes := bytes.Buffer{}

	nsObj := &v1.Namespace{}
	nsObj.SetName(namespace)

	if kymaCRpath != "" {
		data, err := os.ReadFile(kymaCRpath)
		if err != nil {
			return fmt.Errorf("could not read kyma CR file: %w", err)
		}
		yamlBytes.Write(data)
	} else {
		t, err := template.New("yamlBytes").Parse(kymaCRTemplate)
		if err != nil {
			return fmt.Errorf("could not parse Kyma CR template: %w", err)
		}

		if channel == "" {
			channel = "regular"
		}
		data := struct {
			Channel   string
			Sync      bool
			Namespace string
		}{
			Channel:   channel,
			Sync:      false,
			Namespace: namespace,
		}

		if err := t.Execute(&yamlBytes, data); err != nil {
			return fmt.Errorf("could not build Kyma CR: %w", err)
		}
	}

	result := yamlBytes.Bytes()

	if dryRun {
		fmt.Printf("%s---\n", result)
		return nil
	}

	return k8s.Apply(result)
}
