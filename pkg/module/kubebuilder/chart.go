package kubebuilder

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"
)

const (
	chartFile     = "Chart.yaml"
	chartTemplate = `apiVersion: v2
name: {{ .Name }}-operator
description: A Helm chart for the Operator in a Cluster based on a Kustomize Manifest
type: application
version: {{ .Version }}
appVersion: "{{ .Version }}"
`
)

// Add chart generates a Chart.yaml file with the given name and version and saves it to the chartFolder.
func addChart(name, version, chartFolder string) error {
	t, err := template.New("Chart").Parse(chartTemplate)
	if err != nil {
		return err
	}

	data := struct {
		Name    string
		Version string
	}{
		Name:    name,
		Version: version,
	}

	w := &bytes.Buffer{}
	if err := t.Execute(w, data); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(chartFolder, chartFile), w.Bytes(), os.ModePerm)
}
