package system

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/nice"
	"github.com/spf13/cobra"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new kyma CLI command
func NewCmd(o *Options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "system SYSTEM_NAME",
		Short: "Creates a system on the Kyma cluster with the specified name.",
		Long: `Use this command to create a system on the Kyma cluster.

### Detailed description

A system in Kyma is used to bind external systems and applications to the cluster and allow Kyma services and functions to communicate with them and receive events.
Once a system is created in Kyma, use the token provided by this command to allow the external system to access the Kyma cluster.

To generate a new token, rerun the same command with the ` + "`--update`" + ` flag.

`,
		RunE:    func(_ *cobra.Command, args []string) error { return c.Run(args) },
		Aliases: []string{"sys"},
	}

	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Namespace to bind the system to.")
	cmd.Flags().BoolVarP(&o.Update, "update", "u", false, "Updates an existing system and/or generates a new token for it.")
	cmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "", "Specifies the format of the command output. Supported formats: YAML, JSON.")
	cmd.Flags().DurationVarP(&o.Timeout, "timeout", "", 2*time.Minute, "Timeout after which CLI stops watching the installation progress.")

	return cmd
}

func (c *command) Run(args []string) error {
	if c.opts.OutputFormat == "" && !c.opts.NonInteractive {
		// TODO remove when out of alpha
		np := nice.Nice{}
		np.PrintImportant("WARNING: This command is experimental and might change in its final version.")
	}

	var err error
	if c.K8s, err = kube.NewFromConfigWithTimeout("", c.KubeconfigPath, c.opts.Timeout); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	if c.opts.Namespace == "" {
		c.opts.Namespace = c.K8s.DefaultNamespace()
	}

	// validate cluster state
	if _, err := c.K8s.Static().CoreV1().Namespaces().Get(context.Background(), c.opts.Namespace, metav1.GetOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			return fmt.Errorf("Namespace %s does not exist", c.opts.Namespace)
		}
		return errors.Wrapf(err, "Cannot get namespace %s", c.opts.Namespace)
	}

	// create system
	name := args[0]

	c.newStep("Creating system")
	_, err = createSystem(name, c.opts.Update, c.K8s)
	if err != nil {
		c.failStep()
		return errors.Wrap(err, "Could not create System")
	}
	c.successStep("System created")

	//bind NS
	c.newStep("Binding namespace")
	if err := bindNamespace(name, c.opts.Namespace, c.K8s); err != nil {
		c.failStep()
		return err
	}
	c.successStep("Namespace bound")

	// create token
	c.newStep("Generating access token")
	token, err := createToken(name, c.opts.Namespace, c.K8s)
	if err != nil {
		c.failStep()
		return err
	}
	c.successStep("Token generated")

	// print result
	// remove fields that are irrelevant for the consumer:
	unstructured.RemoveNestedField(token.Object, "context")

	unstructured.RemoveNestedField(token.Object, "metadata", "generation")
	unstructured.RemoveNestedField(token.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(token.Object, "metadata", "selfLink")
	unstructured.RemoveNestedField(token.Object, "metadata", "uid")

	switch c.opts.OutputFormat {
	case "yaml":
		b, err := yaml.Marshal(token.Object)
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	case "json":
		b, err := json.MarshalIndent(token.Object, "", " ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))

	default:

		t, exists, err := unstructured.NestedString(token.Object, "status", "token")
		if err != nil {
			return err
		}
		if !exists || t == "" {
			return errors.New("token not found")
		}

		url, exists, err := unstructured.NestedString(token.Object, "status", "url")
		if err != nil {
			return err
		}
		if !exists || url == "" {
			return errors.New("system access URL not found")
		}

		fmt.Println("")
		nicePrint := nice.Nice{
			NonInteractive: c.Factory.NonInteractive,
		}
		fmt.Print("System name:\t\t")
		nicePrint.PrintImportant(name)

		fmt.Print("Namespace:\t\t")
		nicePrint.PrintImportant(c.opts.Namespace)

		fmt.Print("Token:\t\t\t")
		nicePrint.PrintImportant(t)

		fmt.Print("URL:\t\t\t")
		nicePrint.PrintImportant(url)
	}

	return nil
}

// newStep creates a step only if the output format supports it.
func (c *command) newStep(msg string) {
	if c.opts.OutputFormat == "" {
		c.CurrentStep = c.NewStep(msg)
	}
}

// successStep marks a step as successful only if the output format supports it.
func (c *command) successStep(msg string) {
	if c.opts.OutputFormat == "" {
		c.CurrentStep.Successf(msg)
	}
}

// failStep marks a step as failed only if the output format supports it.
func (c *command) failStep() {
	if c.opts.OutputFormat == "" {
		c.CurrentStep.Failure()
	}
}

func createSystem(name string, update bool, k8s kube.KymaKube) (*unstructured.Unstructured, error) {
	sysRes := schema.GroupVersionResource{
		Group:    "applicationconnector.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}
	itm, err := k8s.Dynamic().Resource(sysRes).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(err, "Failed to check System")
		}
	}

	if itm != nil && !update {
		return nil, errors.New("System already exists. To update an existing system or create a new token for an existing system, use the `--update` flag")
	}

	if itm != nil && update {
		// update fields here with "unstructured.SetNestedField()"

		_, err = k8s.Dynamic().Resource(sysRes).Update(context.Background(), itm, metav1.UpdateOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to update system.")
		}
	} else {
		newSys := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "applicationconnector.kyma-project.io/v1alpha1",
				"kind":       "Application",
				"metadata": map[string]interface{}{
					"name": name,
				},
			},
		}

		_, err = k8s.Dynamic().Resource(sysRes).Create(context.Background(), newSys, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create system.")
		}
	}

	// checks if the system resource is in deployed state
	checkFn := func(u *unstructured.Unstructured) (bool, error) {
		status, exists, err := unstructured.NestedString(u.Object, "status", "installationStatus", "status")
		if err != nil {
			return false, errors.Wrap(err, "Error checking system readiness")
		}
		return exists && status == "deployed", nil
	}

	err = k8s.WatchResource(sysRes, name, "", checkFn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to wait for system deployment")
	}

	return k8s.Dynamic().Resource(sysRes).Get(context.Background(), name, metav1.GetOptions{})
}

func bindNamespace(name string, namespace string, k8s kube.KymaKube) error {
	sysMappingRes := schema.GroupVersionResource{
		Group:    "applicationconnector.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "applicationmappings",
	}
	itm, err := k8s.Dynamic().Resource(sysMappingRes).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return errors.Wrap(err, "Failed to check System")
		}
	}

	if itm != nil {
		// always update the mapping
		_, err = k8s.Dynamic().Resource(sysMappingRes).Namespace(namespace).Update(context.Background(), itm, metav1.UpdateOptions{})

	} else {
		newSysMapping := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "applicationconnector.kyma-project.io/v1alpha1",
				"kind":       "ApplicationMapping",
				"metadata": map[string]interface{}{
					"name": name,
				},
			},
		}

		_, err = k8s.Dynamic().Resource(sysMappingRes).Namespace(namespace).Create(context.Background(), newSysMapping, metav1.CreateOptions{})
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to bind the system %s to namespace %s", name, namespace)
	}
	return nil
}

func createToken(name, namespace string, k8s kube.KymaKube) (*unstructured.Unstructured, error) {
	tokenRequestRes := schema.GroupVersionResource{
		Group:    "applicationconnector.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "tokenrequests",
	}
	newToken := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "applicationconnector.kyma-project.io/v1alpha1",
			"kind":       "TokenRequest",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}

	// Check if a token with that name already exists
	itm, err := k8s.Dynamic().Resource(tokenRequestRes).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(err, "Failed to create Token, application does not exist")
		}
	}

	if itm != nil {
		// Token already exists, deleting it.
		err = k8s.Dynamic().Resource(tokenRequestRes).Namespace(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			if !k8sErrors.IsNotFound(err) {
				return nil, errors.Wrap(err, "Failed to remove Token")
			}
		}
	}

	_, err = k8s.Dynamic().Resource(tokenRequestRes).Namespace(namespace).Create(context.Background(), newToken, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create token.")
	}

	err = k8s.WatchResource(tokenRequestRes, name, namespace, func(u *unstructured.Unstructured) (bool, error) {
		itm, err := k8s.Dynamic().Resource(tokenRequestRes).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrap(err, "Failed to request token")
		}

		state, exists, err := unstructured.NestedString(itm.Object, "status", "state")
		return exists && state == "OK", err
	})

	if err != nil {
		return nil, err
	}

	return k8s.Dynamic().Resource(tokenRequestRes).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
}
