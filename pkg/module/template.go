package module

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/kyma-project/cli/pkg/module/oci"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"sigs.k8s.io/yaml"
)

const modTemplate = `apiVersion: operator.kyma-project.io/v1alpha1
kind: ModuleTemplate
metadata:
  name: moduletemplate-{{ .ShortName }}
  namespace: kyma-system
  labels:
    "operator.kyma-project.io/managed-by": "kyma-operator"
    "operator.kyma-project.io/controller-name": "manifest"
    "operator.kyma-project.io/module-name": "{{ .ShortName }}"
  annotations:
    "operator.kyma-project.io/module-version": "{{ .Descriptor.Version }}"
    "operator.kyma-project.io/module-provider": "{{ .Descriptor.ComponentSpec.Provider }}"
    "operator.kyma-project.io/descriptor-schema-version": "{{ .Descriptor.Metadata.Version }}"
spec:
  target: remote
  channel: {{.Channel}}
  data:
{{.Data | indent 4}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
`

func Template(d *v2.ComponentDescriptor, channel, path string, fs vfs.FileSystem) (string, error) {

	ref, err := oci.ParseRef(d.Name)
	if err != nil {
		return "", err
	}

	data, err := vfs.ReadFile(fs, filepath.Join(path, "default.yaml"))
	if err != nil {
		return "", err
	}

	td := struct { // Custom struct for the template
		ShortName  string                  // Last part of the component descriptor name
		Descriptor *v2.ComponentDescriptor // descriptor info for the template
		Channel    string
		Data       string // contents for the spec.data section of the template taken from the defaults.yaml file in the mod folder
	}{
		ShortName:  ref.ShortName(),
		Descriptor: d,
		Channel:    channel,
		Data:       string(data),
	}

	t, err := template.New("modTemplate").Funcs(template.FuncMap{"yaml": yaml.Marshal, "indent": Indent}).Parse(modTemplate)
	if err != nil {
		return "", err
	}

	w := &strings.Builder{}
	if err := t.Execute(w, td); err != nil {
		return "", fmt.Errorf("could not generate a module template out of the component descriptor: %w", err)
	}

	return w.String(), nil
}

// indent prepends the given number of whitespaces to eachline in the given string
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
