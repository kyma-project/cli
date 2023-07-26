package module

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"sigs.k8s.io/yaml"
)

const (
	modTemplate = `apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: {{.ResourceName}}
  namespace: kcp-system
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
  data:
{{- with .Data}}
{{. | indent 4}}
{{- end}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
`
)

func Template(remote ocm.ComponentVersionAccess, moduleTemplateName, channel string, data []byte, labels, annotations map[string]string) ([]byte, error) {
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
	td := struct { // Custom struct for the template
		ResourceName string                              // K8s resource name of the generated ModuleTemplate
		Descriptor   compdesc.ComponentDescriptorVersion // descriptor info for the template
		Channel      string
		Data         string // contents for the spec.data section of the template taken from the defaults.yaml file in the mod folder
		Labels       map[string]string
		Annotations  map[string]string
	}{
		ResourceName: resourceName,
		Descriptor:   cva,
		Channel:      channel,
		Data:         string(data),
		Labels:       labels,
		Annotations:  annotations,
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
