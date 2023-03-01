package module

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	modTemplate = `apiVersion: operator.kyma-project.io/v1beta1
kind: ModuleTemplate
metadata:
  name: moduletemplate-{{ .ShortName }}
  namespace: kcp-system
  labels:
    "operator.kyma-project.io/managed-by": "lifecycle-manager"
    "operator.kyma-project.io/controller-name": "manifest"
    "operator.kyma-project.io/module-name": "{{ .ShortName }}"
spec:
  target: {{.Target}}
  channel: {{.Channel}}
  data:
{{.Data | indent 4}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
`

	//nolint:gosec
	OCIRegistryCredLabel = "oci-registry-cred"
)

func Template(
	remote ocm.ComponentVersionAccess, channel, target string, data []byte, registryCredSelector string,
) ([]byte, error) {
	descriptor := remote.GetDescriptor()
	if registryCredSelector != "" {
		selector, err := metav1.ParseToLabelSelector(registryCredSelector)
		if err != nil {
			return nil, err
		}
		matchLabels, err := json.Marshal(selector.MatchLabels)
		if err != nil {
			return nil, err
		}
		for i := range descriptor.Resources {
			resource := &descriptor.Resources[i]
			resource.SetLabels(
				[]ocmv1.Label{{
					Name:  OCIRegistryCredLabel,
					Value: matchLabels,
				}},
			)
		}
	}
	ref, err := oci.ParseRef(descriptor.Name)
	if err != nil {
		return nil, err
	}

	cva, err := compdesc.Convert(descriptor)
	if err != nil {
		return nil, err
	}

	td := struct { // Custom struct for the template
		ShortName  string                              // Last part of the component descriptor name
		Descriptor compdesc.ComponentDescriptorVersion // descriptor info for the template
		Channel    string
		Target     string
		Data       string // contents for the spec.data section of the template taken from the defaults.yaml file in the mod folder
	}{
		ShortName:  ref.ShortName(),
		Descriptor: cva,
		Channel:    channel,
		Target:     target,
		Data:       string(data),
	}

	t, err := template.New("modTemplate").Funcs(template.FuncMap{"yaml": yaml.Marshal, "indent": Indent}).Parse(modTemplate)
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	if err := t.Execute(w, td); err != nil {
		return nil, fmt.Errorf("could not generate a module template out of the component descriptor: %w", err)
	}

	return w.Bytes(), nil
}

// Indent prepends the given number of whitespaces to eachline in the given string
func Indent(n int, in string) string {
	out := strings.Builder{}

	lines := strings.Split(in, "\n")

	// remove empty line at the end of the file if any
	if len(strings.TrimSpace(lines[len(lines)-1])) == 0 {
		lines = lines[:len(lines)-1]
	}

	for i, line := range lines {
		out.WriteString(strings.Repeat(" ", n))
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}
