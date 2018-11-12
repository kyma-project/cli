package ctx

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kyma-incubator/kymactl/config"
	"github.com/spf13/cobra"
)

var (
	longDesc = `Manage Kyma contexts

A context defines which cluster and configuration will be used by the kyma CLI.

List the available contexts in the configuration or apply the provided context.`

	usageTmpl = `{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Usage:
  kyma ctx [context]
  kyma ctx SUBCOMMAND [options]

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

func NewCmdCtx() *cobra.Command {
	ctxCmd := &cobra.Command{
		Use: "ctx",
		DisableFlagsInUseLine: true,
		Short: "Operate Kyma contexts",
		Long:  longDesc,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				printContexts()
			case 1:
				applyCtx(args[0])
			}
		},
	}

	ctxCmd.SetUsageTemplate(usageTmpl)

	//subcommands
	ctxCmd.AddCommand(newCmdCtxAdd())
	ctxCmd.AddCommand(newCmdCtxRm())

	return ctxCmd
}

func printContexts() {
	fmt.Println("Available contexts:")
	for k, v := range contextConfig().Contexts {
		fmt.Printf("  %s -> %s\n", k, v)
	}
}

func applyCtx(ctxName string) {
	cfg := contextConfig()
	cl, exists := cfg.Contexts[ctxName]
	if !exists {
		fmt.Printf("Context %s not found...Maybe you forgot to add the context to the configuration?\n\nTo Add the context run:\n  kyma ctx add -n %s -url [cluster URL]\n\n", ctxName, ctxName)
		os.Exit(1)
	}

	cmd := exec.Command("kubectl", "config", "use", cl)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Error applying Kubeconfig: %s\n", out)
		os.Exit(1)
	}
	// save current context into config file
	cfg.CTX = ctxName
	if err := config.SaveContext(cfg); err != nil {
		fmt.Printf("Error saving context configuration: %s\n", err)
		os.Exit(1)
	}
}

func contextConfig() *config.ContextConfig {
	// load context config
	cfg, err := config.Context()
	if err != nil {
		fmt.Printf("Error getting context configuration: %s\n", err)
		os.Exit(1)
	}
	return cfg
}
