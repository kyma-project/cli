package provision

import (
	"encoding/json"
	"fmt"
	"os"
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
	parametersPath  string
}

func NewProvisionCMD() *cobra.Command {
	config := provisionConfig{}

	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a Kyma cluster on the BTP.",
		Long: `Use this command to provision a Kyma environment on the SAP BTP platform.
`,
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runProvision(&config))
		},
	}

	cmd.Flags().StringVar(&config.credentialsPath, "credentials-path", "", "Path to the CIS credentials file.")

	cmd.Flags().StringVar(&config.plan, "plan", "trial", "Name of the Kyma environment plan, e.g trial, azure, aws, gcp.")
	cmd.Flags().StringVar(&config.environmentName, "environment-name", "kyma", "Name of the environment in the BTP.")
	cmd.Flags().StringVar(&config.clusterName, "cluster-name", "kyma", "Name of the Kyma cluster.")
	cmd.Flags().StringVar(&config.region, "region", "", "Name of the region of the Kyma cluster.")
	cmd.Flags().StringVar(&config.owner, "owner", "", "Email of the owner of the Kyma cluster.")
	cmd.Flags().StringVar(&config.parametersPath, "parameters", "", "Path to the JSON file with Kyma configuration.")

	_ = cmd.MarkFlagRequired("credentials-path")
	_ = cmd.MarkFlagRequired("owner")
	// mark flag parameters exclusive to clusterName and region
	cmd.MarkFlagsMutuallyExclusive("parameters", "cluster-name")
	cmd.MarkFlagsMutuallyExclusive("parameters", "region")

	return cmd
}

func runProvision(config *provisionConfig) clierror.Error {
	// TODO: is the credentials a good name for this field? it contains much more than credentials only
	credentials, err := auth.LoadCISCredentials(config.credentialsPath)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to load credentials"))
	}

	token, err := auth.GetOAuthToken(
		credentials.GrantType,
		credentials.UAA.URL,
		credentials.UAA.ClientID,
		credentials.UAA.ClientSecret,
	)
	if err != nil {
		var hints []string
		if strings.Contains(err.String(), "Internal Server Error") {
			hints = append(hints, "check if CIS grant type is set to client credentials")
		}

		return clierror.WrapE(err, clierror.New("failed to get access token", hints...))
	}

	// TODO: maybe we should pass only credentials.Endpoints?
	localCISClient := cis.NewLocalClient(credentials, token)

	kymaParameters, err := buildParameters(config)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to prepare kyma parameters"))
	}

	ProvisionEnvironment := &cis.ProvisionEnvironment{
		EnvironmentType: "kyma",
		PlanName:        config.plan,
		Name:            config.environmentName,
		User:            config.owner,
		Parameters:      *kymaParameters,
	}

	response, err := localCISClient.Provision(ProvisionEnvironment)
	if err != nil {
		return clierror.WrapE(err, clierror.New("failed to provision kyma runtime"))
	}

	fmt.Printf("Kyma environment provisioning, environment name: '%s', id: '%s'\n", response.Name, response.ID)

	return nil
}

func buildParameters(config *provisionConfig) (*cis.KymaParameters, clierror.Error) {
	if config.parametersPath == "" {
		return &cis.KymaParameters{
			Name:   config.clusterName,
			Region: config.region,
		}, nil
	}
	return loadParameters(config.parametersPath)
}

func loadParameters(path string) (*cis.KymaParameters, clierror.Error) {
	parametersBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to read parameters file", "Make sure the path to the parameters file is correct."))
	}

	parameters := cis.KymaParameters{}
	err = json.Unmarshal(parametersBytes, &parameters)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to unmarshal file data", "Make sure the parameters file is in the correct format."))
	}

	return &parameters, nil
}
