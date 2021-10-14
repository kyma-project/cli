package store

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new store command
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Imports certificates to local keychain or adds domains to the local host file.",
	}
	return cmd
}
