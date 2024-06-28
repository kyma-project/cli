package hana

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type hanaCredentialsConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	name      string
	namespace string
	user      bool
	password  bool
}

type credentials struct {
	username string
	password string
}

func NewHanaCredentialsCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaCredentialsConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Print credentials of the Hana instance.",
		Long:  "Use this command to print credentials of the Hana instance on the SAP Kyma platform.",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runCredentials(&config))
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")

	cmd.Flags().BoolVar(&config.user, "user", false, "Print only user name")
	cmd.Flags().BoolVar(&config.password, "password", false, "Print only password")

	_ = cmd.MarkFlagRequired("name")
	cmd.MarkFlagsMutuallyExclusive("user", "password")

	return cmd
}

func runCredentials(config *hanaCredentialsConfig) clierror.Error {
	fmt.Printf("Getting Hana credentials (%s/%s).\n", config.namespace, config.name)

	credentials, err := getHanaCredentials(config)
	if err != nil {
		return err
	}
	printCredentials(config, credentials)
	return nil
}

func printCredentials(config *hanaCredentialsConfig, credentials credentials) {
	if config.user {
		fmt.Printf("%s", credentials.username)
	} else if config.password {
		fmt.Printf("%s", credentials.password)
	} else {
		fmt.Printf("Credentials: %s / %s\n", credentials.username, credentials.password)
	}
}

func getHanaCredentials(config *hanaCredentialsConfig) (credentials, clierror.Error) {
	secret, err := config.KubeClient.Static().CoreV1().Secrets(config.namespace).Get(config.Ctx, config.name, metav1.GetOptions{})
	if err != nil {
		return handleGetHanaCredentialsError(err)
	}
	return credentials{
		username: string(secret.Data["username"]),
		password: string(secret.Data["password"]),
	}, nil
}

func handleGetHanaCredentialsError(err error) (credentials, clierror.Error) {
	hints := []string{
		"Make sure that Hana is run and ready to use. You can use command 'kyma hana check'.",
	}

	if err.Error() == "Unauthorized" {
		hints = append(hints, "Make sure that your kubeconfig has access to kubernetes.")
	}

	credErr := clierror.Wrap(err,
		clierror.New("failed to get Hana credentials", hints...),
	)

	return credentials{}, credErr
}
