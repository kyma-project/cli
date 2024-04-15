package hana

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type hanaCredentialsConfig struct {
	*cmdcommon.KymaConfig
	kubeClient kube.Client

	kubeconfig string
	name       string
	namespace  string
	user       bool
	password   bool
}

type credentials struct {
	username string
	password string
}

func NewHanaCredentialsCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaCredentialsConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Print credentials of the Hana instance.",
		Long:  "Use this command to print credentials of the Hana instance on the SAP Kyma platform.",
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runCredentials(&config)
		},
	}

	cmd.Flags().StringVar(&config.kubeconfig, "kubeconfig", "", "Path to the Kyma kubeconfig file.")

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")

	cmd.Flags().BoolVar(&config.user, "user", false, "Print only user name")
	cmd.Flags().BoolVar(&config.password, "password", false, "Print only password")

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("name")
	cmd.MarkFlagsMutuallyExclusive("user", "password")

	return cmd
}

func (pc *hanaCredentialsConfig) complete() error {
	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

func runCredentials(config *hanaCredentialsConfig) error {
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

func getHanaCredentials(config *hanaCredentialsConfig) (credentials, error) {
	secret, err := config.kubeClient.Static().CoreV1().Secrets(config.namespace).Get(config.Ctx, config.name, metav1.GetOptions{})
	if err != nil {
		return handleGetHanaCredentialsError(err)
	}
	return credentials{
		username: string(secret.Data["username"]),
		password: string(secret.Data["password"]),
	}, nil
}

func handleGetHanaCredentialsError(err error) (credentials, error) {
	error := &clierror.Error{
		Message: "failed to get Hana credentails",
		Details: err.Error(),
		Hints: []string{
			"Make sure that Hana is run and ready to use. You can use command 'kyma hana check'.",
		},
	}
	if error.Details == "Unauthorized" {
		error.Hints = []string{
			"Make sure that your kubeconfig has access to kubernetes.",
		}
	}
	return credentials{}, error
}
