package hana

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/btp/hana"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type hanaMapConfig struct {
	*cmdcommon.KymaConfig

	hanaID          string
	credentialsPath string
}

func NewMapHanaCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaMapConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "map",
		Short: "Maps an SAP HANA instance to the Kyma cluster",
		Long:  "Use this command to map an SAP HANA instance to the Kyma cluster.",
		Example: `  # Map SAP Hana DB instance with Kyma runtime
  kyma alpha hana map --credentials-path hana-admin-creds.json`,

		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("credentials-path"),
			))
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runHanaMap(&config))
		},
	}

	cmd.Flags().StringVar(&config.credentialsPath, "credentials-path", "", "Path to the credentials json file")
	cmd.Flags().StringVar(&config.hanaID, "hana-id", "", "SAP HANA instance ID")

	return cmd
}

func runHanaMap(config *hanaMapConfig) clierror.Error {
	client, err := config.GetKubeClientWithClierr()
	if err != nil {
		return err
	}

	clusterID, err := getClusterID(config.Ctx, client.Static())
	if err != nil {
		return clierror.WrapE(err, clierror.New("while getting Kyma cluster ID"))
	}

	credentials, err := hana.ReadCredentialsFromFile(config.credentialsPath)
	if err != nil {
		return clierror.WrapE(err, clierror.New("while reading SAP Hana credentials from file"))
	}

	token, err := auth.GetOAuthToken("client_credentials", credentials.UAA.URL, credentials.UAA.ClientID, credentials.UAA.ClientSecret)
	if err != nil {
		return clierror.WrapE(err, clierror.New("while getting OAuth token"))
	}

	// get Hana ID if not provided by the user
	hanaID := config.hanaID
	if hanaID == "" {
		hanaID, err = hana.GetID(credentials.BaseURL, token.AccessToken)
		if err != nil {
			return clierror.WrapE(err, clierror.New("while getting SAP Hana ID"))

		}
	}
	err = hana.MapInstance(credentials.BaseURL, clusterID, hanaID, token.AccessToken)
	if err != nil {
		return clierror.WrapE(err, clierror.New("while mapping Kyma environment instance with SAP Hana instance"))
	}

	out.Msgfln("Hana with id '%s' is mapped to the cluster with id '%s'", hanaID, clusterID)
	return nil
}

func getClusterID(ctx context.Context, client kubernetes.Interface) (string, clierror.Error) {
	secret, geterr := client.CoreV1().Secrets("kyma-system").Get(ctx, "sap-btp-manager", metav1.GetOptions{})
	if geterr != nil {
		return "", clierror.Wrap(geterr, clierror.New("failed to get secret kyma-system/sap-btp-manager"))
	}

	if secret.Data["cluster_id"] == nil {
		return "", clierror.New("cluster_id not found in the secret kyma-system/sap-btp-manager")
	}

	return string(secret.Data["cluster_id"]), nil
}
