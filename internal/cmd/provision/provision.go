package provision

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/btp"
	"github.com/spf13/cobra"
)

type provisionConfig struct {
	credentialsPath string
	plan            string
	environmentName string
	clusterName     string
	region          string
}

func NewProvisionCMD() *cobra.Command {
	config := provisionConfig{}

	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a Kyma cluster on the BTP.",
		Long: `Use this command to provision a Kyma environment on the SAP BTP platform.
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runProvision(&config)
		},
	}

	cmd.PersistentFlags()
	cmd.Flags().StringVar(&config.credentialsPath, "credentials-path", "", "Path to the CIS credentials file.")

	cmd.Flags().StringVar(&config.plan, "plan", "trial", "Name of the Kyma environment plan, e.g trial, azure, aws, gcp.")
	cmd.Flags().StringVar(&config.environmentName, "environment-name", "kyma", "Name of the environment in the BTP.")
	cmd.Flags().StringVar(&config.clusterName, "cluster-name", "kyma", "Name of the Kyma cluster.")
	cmd.Flags().StringVar(&config.region, "region", "", "Name of the region of the Kyma cluster.")

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

	localCISClient := btp.NewLocalClient(credentials, token)

	ProvisionEnvironment := &btp.ProvisionEnvironment{
		EnvironmentType: "kyma",
		PlanName:        config.plan,
		Name:            config.environmentName,
		User:            "kyma-cli",
		Parameters: btp.KymaParameters{
			Name:   config.clusterName,
			Region: config.region,
		},
	}
	response, err := localCISClient.Provision(ProvisionEnvironment)
	if err != nil {
		return fmt.Errorf("failed to provision kyma runtime: %s", err.Error())
	}

	fmt.Printf("Kyma environment provisioning, environment name: '%s', id: '%s'\n", response.Name, response.ID)

	return nil
}
