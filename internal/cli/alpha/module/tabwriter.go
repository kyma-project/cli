package module

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
)

const (
	TemplateStateAnnotation = "state.cmd.kyma-project.io"
	TemplateModuleNameLabel = v1beta1.ModuleName
)

type TemplateTable interface {
	Print(writer io.Writer, templates []v1beta1.ModuleTemplate) error
}

var _ TemplateTable = &DefaultTemplateTable{}

func NewDefaultTemplateTable(noHeaders, appendState bool) *DefaultTemplateTable {
	entries := templateTableEntries{
		{
			name: "Template",
			valueFn: func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string {
				return fmt.Sprintf("%s/%s", template.GetNamespace(), template.GetName())
			},
		},
		{
			name: TemplateModuleNameLabel,
			valueFn: func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string {
				return template.GetLabels()[TemplateModuleNameLabel]
			},
		},
		{
			name: "Domain Name (FQDN)",
			valueFn: func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string {
				return descriptor.GetName()
			},
		},
		{
			name: "Channel",
			valueFn: func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string {
				return template.Spec.Channel
			},
		},
		{
			name: "Version",
			valueFn: func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string {
				return descriptor.GetVersion()
			},
		},
	}
	if appendState {
		entries = append(entries, templateTableEntry{
			name: "State",
			valueFn: func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string {
				return template.GetAnnotations()[TemplateStateAnnotation]
			},
		})
	}
	return &DefaultTemplateTable{
		noHeaders:            noHeaders,
		templateTableEntries: entries,
	}
}

type DefaultTemplateTable struct {
	noHeaders bool
	templateTableEntries
}

func (t *DefaultTemplateTable) Print(writer io.Writer, templates []v1beta1.ModuleTemplate) error {
	separator := '\t'
	tabWriter := tabwriter.NewWriter(writer, 0, 8, 2, byte(separator), 0)
	if !t.noHeaders {
		header := strings.TrimSuffix(strings.Join(t.templateTableEntries.headerNames(), string(separator)), string(separator))
		if _, err := tabWriter.Write([]byte(header + "\n")); err != nil {
			return err
		}
	}
	for i := range templates {
		values, err := t.templateTableEntries.values(&templates[i])
		if err != nil {
			return err
		}
		_, _ = tabWriter.Write([]byte(fmt.Sprintf(strings.TrimSuffix(strings.Repeat("%s"+string(separator), len(t.templateTableEntries)), string(separator))+"\n", values...)))
	}
	return tabWriter.Flush()
}

type templateTableEntries []templateTableEntry
type templateTableEntry struct {
	name    string
	valueFn func(template *v1beta1.ModuleTemplate, descriptor *v1beta1.Descriptor) string
}

func (e templateTableEntries) headerNames() []string {
	names := make([]string, 0, len(e))
	for i := range e {
		names = append(names, e[i].name)
	}
	return names
}
func (e templateTableEntries) values(template *v1beta1.ModuleTemplate) ([]any, error) {
	descriptor, err := template.Spec.GetDescriptor()
	if err != nil {
		return nil, err
	}
	values := make([]any, 0, len(e))
	for i := range e {
		values = append(values, e[i].valueFn(template, descriptor))
	}
	return values, nil
}
