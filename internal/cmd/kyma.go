package cmd

import (
	"context"
	"github.com/kyma-project/cli.v3/internal/cmd/modules"

	"github.com/kyma-project/cli.v3/internal/cmd/access"

	"github.com/kyma-project/cli.v3/internal/cmd/hana"
	"github.com/kyma-project/cli.v3/internal/cmd/imageimport"
	"github.com/kyma-project/cli.v3/internal/cmd/oidc"
	"github.com/kyma-project/cli.v3/internal/cmd/provision"
	"github.com/kyma-project/cli.v3/internal/cmd/referenceinstance"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewKymaCMD() *cobra.Command {
	config := &cmdcommon.KymaConfig{
		Ctx: context.Background(),
	}

	cmd := &cobra.Command{
		Use: "kyma",

		// Affects children as well
		// by default Cobra adds `Error:` to the front of the error message, we want to supress it
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := cmd.Help(); err != nil {
				_ = err
			}
		},
	}

	cmd.AddCommand(hana.NewHanaCMD(config))
	cmd.AddCommand(imageimport.NewImportCMD(config))
	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(referenceinstance.NewReferenceInstanceCMD(config))
	cmd.AddCommand(access.NewAccessCMD(config))
	cmd.AddCommand(oidc.NewOIDCCMD(config))
	cmd.AddCommand(modules.NewModulesCMD())

	return cmd
}
