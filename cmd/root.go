package cmd

import (
	"fmt"
	"os"

	"github.com/kyma-incubator/kymactl/cmd/ctx"

	"github.com/spf13/cobra"
)

// Root is the entry point of the Kyma CLI tool
var rootCmd = &cobra.Command{
	Use:   "kyma",
	Short: "Kyma is a cloud extension platform for SAP Commerce products",
	Long: `Kyma:
	Kyma is a Cloud Native plaform.
	It uses latest tech and frameworks to build stuff.
	Coolness guaranteed.`,
}

// global flags
var (
	Verbose bool
)

func init() {
	//flags
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	// rootCmd.PersistentFlags().StringP("author", "a", "Borja Clemente", "Author name")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.SetDefault("author", "Borja Clemente <borja.clemente.castanera@sap.com>")

	// subcommands
	rootCmd.AddCommand(ctx.NewCmdCtx())
	rootCmd.AddCommand(newCmdStatus())
	rootCmd.AddCommand(newCmdVersion())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
