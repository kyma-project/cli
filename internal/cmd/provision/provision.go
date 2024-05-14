package provision

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/btp/cis"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/spf13/cobra"
)

type provisionConfig struct {
	credentialsPath string
	plan            string
	environmentName string
	clusterName     string
	region          string
	owner           string
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

	cmd.PersistentFlags() // TODO: do we need this?
	cmd.Flags().StringVar(&config.credentialsPath, "credentials-path", "", "Path to the CIS credentials file.")

	cmd.Flags().StringVar(&config.plan, "plan", "trial", "Name of the Kyma environment plan, e.g trial, azure, aws, gcp.")
	cmd.Flags().StringVar(&config.environmentName, "environment-name", "kyma", "Name of the environment in the BTP.")
	cmd.Flags().StringVar(&config.clusterName, "cluster-name", "kyma", "Name of the Kyma cluster.")
	cmd.Flags().StringVar(&config.region, "region", "", "Name of the region of the Kyma cluster.")
	cmd.Flags().StringVar(&config.owner, "owner", "", "Email of the owner of the Kyma cluster.")

	_ = cmd.MarkFlagRequired("credentials-path")
	_ = cmd.MarkFlagRequired("owner")

	return cmd
}

func runProvision(config *provisionConfig) error {
	// TODO: is the credentials a good name for this field? it contains much more than credentials only
	credentials, err := auth.LoadCISCredentials(config.credentialsPath)
	if err != nil {
		return clierror.Wrap(err, &clierror.Error{Message: "failed to load credentials"})
	}

	token, err := auth.GetOAuthToken(
		credentials.GrantType,
		credentials.UAA.URL,
		credentials.UAA.ClientID,
		credentials.UAA.ClientSecret,
	)
	if err != nil {
		hints := []string{}
		if strings.Contains(err.Error(), "Internal Server Error") {
			hints = append(hints, "check if CIS grant type is set to client credentials")
		}

		return clierror.Wrap(err, &clierror.Error{
			Message: "failed to get access token",
			Hints:   hints,
		})
	}

	// TODO: maybe we should pass only credentials.Endpoints?
	localCISClient := cis.NewLocalClient(credentials, token)

	ProvisionEnvironment := &cis.ProvisionEnvironment{
		EnvironmentType: "kyma",
		PlanName:        config.plan,
		Name:            config.environmentName,
		User:            config.owner,
		Parameters: cis.KymaParameters{
			Name:   config.clusterName,
			Region: config.region,
		},
	}
	response, err := localCISClient.Provision(ProvisionEnvironment)
	if err != nil {
		return clierror.Wrap(err, &clierror.Error{Message: "failed to provision kyma runtime"})
	}

	fmt.Printf("Kyma environment provisioning, environment name: '%s', id: '%s'\n", response.Name, response.ID)

	return nil
}
