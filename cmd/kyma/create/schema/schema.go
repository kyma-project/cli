package schema

import (
	"errors"
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new kyma CLI command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "schema SCHEMA_NAME",
		Short: "Generates json schema for given name",
		Long:  `Generates json schema for given name`,
		RunE:  func(_ *cobra.Command, args []string) error { return c.Run(args) },
	}

	cmd.Args = cobra.ExactArgs(1)

	return cmd
}

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
)

func (c *command) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: schema name not provided", ErrInvalidArgument)
	}

	schema := args[0]

	reflectSchema, ok := c.opts.RefMap[schema]
	if !ok {
		return fmt.Errorf("%w: %q", ErrNotFound, schema)
	}

	bytes, err := reflectSchema()
	if err != nil {
		return err
	}

	if _, err := c.opts.Write(bytes); err != nil {
		return err
	}

	return nil
}
