package schema

import (
	"io"

	"github.com/kyma-project/cli/internal/cli"
)

type Options struct {
	*cli.Options
	io.Writer
	RefMap map[string]func() ([]byte, error)
}

func NewOptions(o *cli.Options, w io.Writer, m map[string]func() ([]byte, error)) *Options {
	return &Options{
		Options: o,
		Writer:  w,
		RefMap:  m,
	}
}
