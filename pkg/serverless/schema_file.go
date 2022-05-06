package serverless

import (
	"io"

	"github.com/kyma-project/hydroform/function/pkg/workspace"
)

type Schema struct{}

func (s Schema) Write(w io.Writer, _ interface{}) error {
	b, err := workspace.ReflectSchema()
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (s Schema) FileName() string {
	return "schema.json"
}
