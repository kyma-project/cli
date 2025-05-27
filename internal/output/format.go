package output

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	JSONFormat    Format = "json"
	YAMLFormat    Format = "yaml"
	DefaultFormat Format = ""
)

var (
	availableFormats = []Format{
		JSONFormat,
		YAMLFormat,
		DefaultFormat,
	}
)

type Format string

func (f *Format) UnmarshalYAML(unmarshal func(interface{}) error) error {
	outputFormat := ""
	err := unmarshal(&outputFormat)
	if err != nil {
		return errors.Wrap(err, "while unmarshalling quantity")
	}

	return f.Set(outputFormat)
}

func (f *Format) String() string {
	return string(*f)
}

func (f *Format) Set(v string) error {
	for _, format := range availableFormats {
		if *f == format {
			*f = Format(v)
			return nil
		}
	}

	return errors.New(fmt.Sprintf("invalid output format '%s'", *f))
}

func (f *Format) Type() string {
	return "string"
}
