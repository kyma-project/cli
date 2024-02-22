package provision

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/btp"
	"github.com/spf13/cobra"
)

type provisionConfig struct {
	credentialsPath string
}

func NewProvisionCMD() *cobra.Command {
	config := provisionConfig{}

	cmd := &cobra.Command{
		Use: "provision",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runProvision(&config)
		},
	}

	cmd.PersistentFlags()
	cmd.Flags().StringVar(&config.credentialsPath, "credentials-path", "", "Path to the CIS credentials file.")
	_ = cmd.MarkFlagRequired("credentials-path")

	return cmd
}

func runProvision(config *provisionConfig) error {
	credentials, err := btp.LoadCISCredentials(config.credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to load credentials from '%s' file: %s", config.credentialsPath, err.Error())
	}

	token, err := btp.GetOAuthToken(credentials)
	if err != nil {
		return fmt.Errorf("failed to get access token: %s", err.Error())
	}

	// TODO: remove in next interation
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", data)

	return nil
}
