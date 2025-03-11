package kubeconfig

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

type generateConfig struct {
	*cmdcommon.KymaConfig

	// common flags
	output string

	// ServiceAccount-based flow flags
	serviceAccount string
	clusterRole    string
	namespace      string
	time           string
	permanent      bool
}

func newGenerateCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := &generateConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use: "generate",
		Example: `# generate a kubeconfig with a ServiceAccount-based token and certificate
  kyma@v3 alpha kubeconfig generate --serviceaccount <sa_name> --clusterrole <cr_name> --namespace <ns_name> --permanent

# generate a kubeconfig with an OIDC token
  kyma@v3 alpha kubeconfig generate --token <token>

# generate a kubeconfig with an requested OIDC token
  kyma@v3 alpha kubeconfig generate --id-token-request-url <url>`,
		Short: "Generate kubeconfig with a Service Account-based or oidc tokens",
		Long:  "Use this command to generate kubeconfig file with a Service Account-based or oidc tokens",
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(runGenerate(cfg))
		},
	}

	// common
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the kubeconfig file output. If not provided, the kubeconfig will be printed")

	// ServiceAccount-based flow
	cmd.Flags().StringVar(&cfg.serviceAccount, "serviceaccount", "", "Name of the Service Account to be created")
	cmd.Flags().StringVar(&cfg.clusterRole, "clusterrole", "", "Name of the cluster role to bind the Service Account to")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace in which the resource is created")
	cmd.Flags().StringVar(&cfg.time, "time", "1h", "Determines how long the token should be valid, by default 1h (use h for hours and d for days)")
	cmd.Flags().BoolVar(&cfg.permanent, "permanent", false, "Determines if the token is valid indefinitely")

	// TODO: OIDC flow

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("clusterrole")

	return cmd
}

func runGenerate(cfg *generateConfig) clierror.Error {
	// Create objects
	clierr := createObjects(cfg)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to create objects"))
	}

	// Fill kubeconfig
	generatedKubeconfig, clierr := kubeconfig.Prepare(cfg.Ctx, cfg.KubeClient, cfg.serviceAccount, cfg.namespace, cfg.time, cfg.output, cfg.permanent)
	if clierr != nil {
		return clierr
	}

	// Print or write to file
	if cfg.output != "" {
		err := kube.SaveConfig(generatedKubeconfig, cfg.output)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to save kubeconfig"))
		}
	} else {
		message, err := clientcmd.Write(*generatedKubeconfig)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to print kubeconfig"))
		}
		fmt.Println(string(message))
	}
	return nil
}

func createObjects(cfg *generateConfig) clierror.Error {
	// Create Service Account
	err := resources.CreateServiceAccount(cfg.Ctx, cfg.KubeClient, cfg.serviceAccount, cfg.namespace)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Service Account"))
	}
	// Create Role Binding for the Service Account
	err = resources.CreateClusterRoleBinding(cfg.Ctx, cfg.KubeClient, cfg.serviceAccount, cfg.namespace, cfg.clusterRole)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Cluster Role Binding"))
	}
	// Create a service-account-token type secret
	if cfg.permanent {
		err = resources.CreateServiceAccountToken(cfg.Ctx, cfg.KubeClient, cfg.serviceAccount, cfg.namespace)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create secret"))
		}
	}
	return nil
}
