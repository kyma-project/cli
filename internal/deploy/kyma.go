package deploy

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/kyma-project/cli/internal/kube"
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

// Kyma deploys the Kyma CR
// TODO delete deploy.go when the old reconciler is gone.
func Kyma(k8s kube.KymaKube, dryRun bool) error {
	t, err := template.New("kymaCR").Parse(kymaCRTemplate)
	if err != nil {
		return fmt.Errorf("could not parse Kyma CR template: %w", err)
	}

	data := struct {
		Channel string
		Sync    bool
	}{
		Channel: "stable",
		Sync:    false,
	}

	kymaCR := bytes.Buffer{}
	if err := t.Execute(&kymaCR, data); err != nil {
		return fmt.Errorf("could not build Kyma CR: %w", err)
	}
	if dryRun {
		fmt.Printf("%s\n---\n", kymaCR.String())
		return nil
	}
	return k8s.Apply(kymaCR.Bytes())
}
