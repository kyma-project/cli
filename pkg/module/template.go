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
)

func Template(remote ocm.ComponentVersionAccess, channel, target string, data []byte) ([]byte, error) {
	descriptor := remote.GetDescriptor()
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
