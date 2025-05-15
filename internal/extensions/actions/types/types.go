package types

import (
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
)

const (
	OutputFormatJSON    OutputFormat = "json"
	OutputFormatYAML    OutputFormat = "yaml"
	OutputFormatDefault OutputFormat = ""
)

type OutputFormat string

func (f *OutputFormat) UnmarshalYAML(unmarshal func(interface{}) error) error {
	outputFormat := ""
	err := unmarshal(&outputFormat)
	if err != nil {
		return errors.Wrap(err, "while unmarshalling quantity")
	}

	for _, format := range []OutputFormat{OutputFormatJSON, OutputFormatYAML, OutputFormatDefault} {
		if format == OutputFormat(outputFormat) {
			*f = format
			return nil
		}
	}

	return errors.Newf("invalid output format '%s'", outputFormat)
}
