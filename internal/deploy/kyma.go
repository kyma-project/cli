package deploy

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/kyma-project/cli/internal/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const kymaCRTemplate = `apiVersion: operator.kyma-project.io/v1alpha1
kind: Kyma
metadata:
  annotations:
    cli.kyma-project.io/source: deploy
  labels:
    operator.kyma-project.io/managed-by: lifecycle-manager
  name: default-kyma
  namespace: kcp-system
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
func Kyma(k8s kube.KymaKube, channel, kymaCRpath string, dryRun bool) error {
	// TODO delete deploy.go when the old reconciler is gone.
	kymaCR := bytes.Buffer{}

	if kymaCRpath != "" {
		data, err := os.ReadFile(kymaCRpath)
		if err != nil {
			return fmt.Errorf("could not read kyma CR file: %w", err)
		}
		kymaCR.Write(data)
	} else {
		t, err := template.New("kymaCR").Parse(kymaCRTemplate)
		if err != nil {
			return fmt.Errorf("could not parse Kyma CR template: %w", err)
		}

		if channel == "" {
			channel = "regular"
		}
		data := struct {
			Channel string
			Sync    bool
		}{
			Channel: channel,
			Sync:    false,
		}

		if err := t.Execute(&kymaCR, data); err != nil {
			return fmt.Errorf("could not build Kyma CR: %w", err)
		}
	}
	if dryRun {
		fmt.Printf("%s\n---\n", kymaCR.String())
		return nil
	}
	return k8s.Apply(kymaCR.Bytes())
}
