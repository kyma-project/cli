package extensions

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KymaExtensionsConfig struct {
	extensionsErrors []error
}

func NewBuilder() *KymaExtensionsConfig {
	return &KymaExtensionsConfig{}
}

func AddCmdPersistentFlags(cmd *cobra.Command) {
	// these flags are not operational. it's only to print the help description, and the help cobra with validation
	_ = cmd.PersistentFlags().Bool("skip-extensions", false, "Skip fetching extensions from the cluster")
	_ = cmd.PersistentFlags().Bool("show-extensions-error", false, "Prints a possible error when fetching extensions fails")
}

func (kec *KymaExtensionsConfig) DisplayWarnings(warningWriter io.Writer) {
	if isSubRootCommandUsed("help", "completion", "version") {
		// skip if one of restricted flags is used
		return
	}

	if len(kec.extensionsErrors) > 0 && getBoolFlagValue("--show-extensions-error") {
		// print error as warning if expected and continue
		fmt.Fprintf(warningWriter, "Extensions Warning:\n%s\n\n", errors.NewList(kec.extensionsErrors...).Error())
	} else if len(kec.extensionsErrors) > 0 {
		fmt.Fprintf(warningWriter, "Extensions Warning:\nfailed to fetch all extensions from the cluster. Use the '--show-extensions-error' flag to see more details.\n\n")
	}
}

// build extensions based on extensions configmaps from a cluster
// any errors can be displayed by using the DisplayExtensionsErrors func
func (kec *KymaExtensionsConfig) Build(parentCmd *cobra.Command, kymaConfig *cmdcommon.KymaConfig, availableActions types.ActionsMap) {
	if getBoolFlagValue("--skip-extensions") {
		// skip extensions fetching
		return
	}

	configmapExtensions, err := loadCommandExtensionsFromCluster(kymaConfig.Ctx, kymaConfig.KubeClientConfig)
	if err != nil {
		// set extensionsError and stop
		kec.extensionsErrors = append(kec.extensionsErrors, err)
		return
	}

	for _, cmExt := range configmapExtensions {
		// default
		cmExt.Extension.Default()

		// validate
		err := cmExt.Extension.Validate(availableActions)
		if err != nil {
			kec.extensionsErrors = append(kec.extensionsErrors,
				errors.Wrapf(err, "failed to validate extension from configmap '%s/%s'", cmExt.ConfigMapNamespace, cmExt.ConfigMapName))
			continue
		}

		// build final commands tree
		command, err := buildCommand(cmExt.Extension, availableActions)
		if err != nil {
			kec.extensionsErrors = append(kec.extensionsErrors,
				errors.Wrapf(err, "failed to build extension from configmap '%s/%s'", cmExt.ConfigMapNamespace, cmExt.ConfigMapName))
			continue
		}

		// check command duplicates
		if hasCommand(parentCmd, command) {
			kec.extensionsErrors = append(kec.extensionsErrors,
				errors.Newf("failed to add extension from configmap '%s/%s': base command with name='%s' already exists",
					cmExt.ConfigMapNamespace, cmExt.ConfigMapName, command.Name()))
			continue
		}

		// append extension command
		parentCmd.AddCommand(command)
	}
}

func hasCommand(base *cobra.Command, cmd *cobra.Command) bool {
	cmds := base.Commands()
	for i := range cmds {
		if cmds[i].Name() == cmd.Name() {
			return true
		}
	}

	return false
}

func loadCommandExtensionsFromCluster(ctx context.Context, clientConfig *cmdcommon.KubeClientConfig) ([]types.ConfigmapCommandExtension, error) {
	var cms, cmsError = listCommandExtenionConfigMaps(ctx, clientConfig)
	if cmsError != nil {
		return nil, cmsError
	}

	extensions := []types.ConfigmapCommandExtension{}
	var parseErrors []error
	for _, cm := range cms.Items {
		commandExtension, err := parseRequiredField[types.Extension](cm.Data, types.ExtensionCMDataKey)
		if err != nil {
			parseErrors = append(parseErrors,
				errors.Wrapf(err, "failed to parse configmap '%s/%s'", cm.GetNamespace(), cm.GetName()))
			continue
		}

		if slices.ContainsFunc(extensions, func(e types.ConfigmapCommandExtension) bool {
			return e.Extension.Metadata.Name == commandExtension.Metadata.Name
		}) {
			parseErrors = append(parseErrors,
				errors.Newf("failed to validate configmap '%s/%s': extension with name='%s' already exists",
					cm.GetNamespace(), cm.GetName(), commandExtension.Metadata.Name))
			continue
		}

		extensions = append(extensions, types.ConfigmapCommandExtension{
			ConfigMapName:      cm.GetName(),
			ConfigMapNamespace: cm.GetNamespace(),
			Extension:          *commandExtension,
		})
	}

	return extensions, errors.NewList(parseErrors...)
}

func listCommandExtenionConfigMaps(ctx context.Context, clientConfig *cmdcommon.KubeClientConfig) (*v1.ConfigMapList, error) {
	client, clientErr := clientConfig.GetKubeClient()
	if clientErr != nil {
		return nil, clientErr
	}

	labelSelector := fmt.Sprintf("%s==%s", types.ExtensionCMLabelKey, types.ExtensionCMLabelValue)
	cms, err := client.Static().CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load ConfigMaps from cluster with label %s", labelSelector)
	}

	return cms, nil
}

func parseRequiredField[T any](cmData map[string]string, cmKey string) (*T, error) {
	dataBytes, ok := cmData[cmKey]
	if !ok {
		return nil, errors.Newf("missing .data.%s field", cmKey)
	}

	var data T
	err := yaml.Unmarshal([]byte(dataBytes), &data)
	return &data, err
}

// search os.Args manually to find if user pass given flag and return its value
func getBoolFlagValue(flag string) bool {
	for i, arg := range os.Args {
		//example: --show-extensions-error true
		if arg == flag && len(os.Args) > i+1 {

			value, err := strconv.ParseBool(os.Args[i+1])
			if err == nil {
				return value
			}
		}

		// example: --show-extensions-error or --show-extensions-error=true
		if strings.HasPrefix(arg, flag) && !strings.Contains(arg, "false") {
			return true
		}
	}

	return false
}

// checks if one of given args is on the 2 possition of os.Args (first sub-command)
func isSubRootCommandUsed(args ...string) bool {
	for _, arg := range args {
		if slices.Contains(os.Args, arg) {
			return true
		}
	}

	return false
}
