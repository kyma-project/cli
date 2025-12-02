package provision

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/btp/cis"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/out"
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
		Use:   "provision [flags]",
		Short: "Provisions a Kyma cluster on SAP BTP",
		Long:  `Use this command to provision a Kyma environment on SAP BTP.`,
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("credentials-path", "owner"),
				flags.MarkExclusive("parameters", "cluster-name", "region"),
			))
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runProvision(&config))
		},
	}

	cmd.Flags().StringVar(&config.credentialsPath, "credentials-path", "", "Path to the CIS credentials file")

	cmd.Flags().StringVar(&config.plan, "plan", "trial", "Name of the Kyma environment plan, e.g trial, azure, aws, gcp")
	cmd.Flags().StringVar(&config.environmentName, "environment-name", "kyma", "Name of the SAP BTP environment")
	cmd.Flags().StringVar(&config.clusterName, "cluster-name", "kyma", "Name of the Kyma cluster")
	cmd.Flags().StringVar(&config.region, "region", "", "Name of the region of the Kyma cluster")
	cmd.Flags().StringVar(&config.owner, "owner", "", "Email of the Kyma cluster owner")
	cmd.Flags().StringVar(&config.parametersPath, "parameters", "", "Path to the JSON file with Kyma configuration")

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
		return clierror.WrapE(err, clierror.New("failed to prepare Kyma parameters"))
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
		return clierror.WrapE(err, clierror.New("failed to provision Kyma runtime"))
	}

	out.Msgfln("Kyma environment provisioning, environment name: '%s', id: '%s'", response.Name, response.ID)

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
