package module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/cli/pkg/module/oci"
)

const modTemplate = `apiVersion: operator.kyma-project.io/v1alpha1
kind: ModuleTemplate
metadata:
  name: moduletemplate-{{ .ShortName }}
  namespace: kyma-system
  labels:
    "operator.kyma-project.io/managed-by": "lifecycle-manager"
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

func Template(archive *ctf.ComponentArchive, channel, path string, fs vfs.FileSystem) ([]byte, error) {
	d, err := remoteDescriptor(archive)
	if err != nil {
		return nil, err
	}

	ref, err := oci.ParseRef(d.Name)
	if err != nil {
		return nil, err
	}

	data, err := vfs.ReadFile(fs, filepath.Join(path, "default.yaml"))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	w := &bytes.Buffer{}
	if err := t.Execute(w, td); err != nil {
		return nil, fmt.Errorf("could not generate a module template out of the component descriptor: %w", err)
	}

	return w.Bytes(), nil
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

// remoteDescriptor generates the remote component descriptor from the local archive the same way module.Push does.
// the module template uses the remote descriptor to install the module
func remoteDescriptor(archive *ctf.ComponentArchive) (*v2.ComponentDescriptor, error) {
	store := oci.NewInMemoryCache()

	manifest, err := cdoci.NewManifestBuilder(store, archive).Build(context.TODO())
	if err != nil {
		return nil, err
	}

	r, err := store.Get(manifest.Layers[0])
	if err != nil {
		return nil, err
	}

	data, err := cdoci.ReadComponentDescriptorFromTar(r)
	if err != nil {
		return nil, err
	}

	remoteDesc := &v2.ComponentDescriptor{}
	if err := json.Unmarshal(data, remoteDesc); err != nil {
		return nil, err
	}

	return remoteDesc, nil
}
