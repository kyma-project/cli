package module

import (
	"bytes"
	"fmt"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"strings"
	"text/template"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/cli/pkg/module/oci"
)

const (
	modTemplate = `apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: {{.ResourceName}}
  namespace: {{.Namespace}}
{{- with .Labels}}
  labels:
    {{- range $key, $value := . }}
    {{ printf "%q" $key }}: {{ printf "%q" $value }}
    {{- end}}
{{- end}} 
{{- with .Annotations}}
  annotations:
    {{- range $key, $value := . }}
    {{ printf "%q" $key }}: {{ printf "%q" $value }}
    {{- end}}
{{- end}} 
spec:
  channel: {{.Channel}}
{{- with .CustomStateChecks}}
  customStateCheck:
    {{- range .}}
    - jsonPath: '{{.JSONPath}}'
      value: '{{.Value}}'
      mappedState: '{{.MappedState}}'
    {{- end}}
{{- end}} 
  data:
{{- with .Data}}
{{. | indent 4}}
{{- end}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
`
)

type moduleTemplateData struct {
	ResourceName      string // K8s resource name of the generated ModuleTemplate
	Namespace         string
	Descriptor        compdesc.ComponentDescriptorVersion // descriptor info for the template
	Channel           string
	Data              string // contents for the spec.data section of the template taken from the defaults.yaml file in the mod folder
	Labels            map[string]string
	Annotations       map[string]string
	CustomStateChecks []v1beta2.CustomStateCheck
}

func Template(remote ocm.ComponentVersionAccess, moduleTemplateName, namespace, channel string, data []byte,
	labels, annotations map[string]string, customsStateChecks []v1beta2.CustomStateCheck) ([]byte, error) {
	descriptor := remote.GetDescriptor()
	ref, err := oci.ParseRef(descriptor.Name)
	if err != nil {
		return nil, err
	}

	cva, err := compdesc.Convert(descriptor)
	if err != nil {
		return nil, err
	}

	shortName := ref.ShortName()
	labels["operator.kyma-project.io/module-name"] = shortName
	resourceName := moduleTemplateName
	if len(resourceName) == 0 {
		resourceName = shortName + "-" + channel
	}
	td := moduleTemplateData{
		ResourceName:      resourceName,
		Namespace:         namespace,
		Descriptor:        cva,
		Channel:           channel,
		Data:              string(data),
		Labels:            labels,
		Annotations:       annotations,
		CustomStateChecks: customsStateChecks,
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

// Indent prepends the given number of whitespaces to each line in the given string
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
