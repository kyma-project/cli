package extension

import (
	"fmt"

	"github.com/spf13/cobra"
)

func BuildCommands(extensions []Extension) []*cobra.Command {
	cmds := make([]*cobra.Command, len(extensions))

	for i, extension := range extensions {
		cmds[i] = buildCommandFromExtension(&extension)
	}

	return cmds
}

func buildCommandFromExtension(extension *Extension) *cobra.Command {
	cmd := &cobra.Command{
		Use:   extension.RootCommand.Name,
		Short: extension.RootCommand.Description,
		Long:  extension.RootCommand.DescriptionLong,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := cmd.Help(); err != nil {
				_ = err
			}
		},
	}

	if extension.GenericCommands != nil {
		addGenericCommands(cmd, extension.GenericCommands)
	}

	return cmd
}

func addGenericCommands(cmd *cobra.Command, genericCommands *GenericCommands) {
	if genericCommands.ExplainCommand != nil {
		cmd.AddCommand(buildExplainCommand(genericCommands.ExplainCommand))
	}
}

func buildExplainCommand(explainCommand *ExplainCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: explainCommand.Description,
		Long:  explainCommand.DescriptionLong,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(explainCommand.ExplainOutput)
		},
	}
}
