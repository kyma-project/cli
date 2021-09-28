package undeploy

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sRetry "k8s.io/client-go/util/retry"
	"sync"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"

	"github.com/spf13/cobra"
)

var (
	crdGvr = schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
)

type command struct {
	opts *Options
	cli.Command
	apixClient apixv1beta1client.ApiextensionsV1beta1Interface
}

//NewCmd creates a new kyma command
func NewCmd(o *Options) *cobra.Command {

	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:   "undeploy",
		Short: "Removes Kyma from a running Kubernetes cluster.",
		Long:  `Use this command to clean up Kyma from a running Kubernetes cluster.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return cmd.Run() },
	}

	cobraCmd.Flags().BoolVarP(&o.KeepCRDs, "keep-crds", "", false, "Set --keep-crds=true to keep CRDs on clean-up")
	return cobraCmd
}

//Run runs the command
func (cmd *command) Run() error {
	var err error

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}

	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Cannot initialize the Kubernetes client. Make sure your kubeconfig is valid")
	}

	cmd.apixClient, _ = apixv1beta1client.NewForConfig(cmd.K8s.RestConfig())
	if err != nil {
		return err
	}

	err = cmd.undeployKyma()
	if err != nil {
		return err
	}

	fmt.Println("Kyma successfully removed.")

	return nil
}

func (cmd *command) undeployKyma() error {
	err := cmd.removeFinalizers()
	if err != nil {
		return err
	}

	err = cmd.deleteKymaNamespaces()
	if err != nil {
		return err
	}

	if !cmd.opts.KeepCRDs {
		return cmd.deleteKymaCRDs()
	}

	return nil
}

func (cmd *command) removeFinalizers() error {
	if err := cmd.removeServerlessCredentialsFinalizers(); err != nil {
		return err
	}

	if err := cmd.removeCustomResourcesFinalizers(); err != nil {
		return err
	}

	return nil
}

func (cmd *command) removeServerlessCredentialsFinalizers() error {
	secrets, err := cmd.K8s.Static().CoreV1().Secrets(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{LabelSelector: "serverless.kyma-project.io/config=credentials"})
	if err != nil && !apierr.IsNotFound(err) {
		return err
	}

	if secrets == nil {
		return nil
	}

	for i := range secrets.Items {
		secret := secrets.Items[i]

		if len(secret.GetFinalizers()) <= 0 {
			continue
		}

		secret.SetFinalizers(nil)
		if _, err := cmd.K8s.Static().CoreV1().Secrets(secret.GetNamespace()).Update(context.Background(), &secret, metav1.UpdateOptions{}); err != nil {
			return err
		}
		fmt.Printf("Deleted finalizer from \"%s\" Secret", secret.GetName())
	}

	return nil
}

func (cmd *command) removeCustomResourcesFinalizers() error {
	crds, err := cmd.apixClient.CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{LabelSelector: "reconciler.kyma-project.io/managed-by=reconciler"})
	if err != nil && !apierr.IsNotFound(err) {
		return err
	}

	if crds == nil {
		return nil
	}

	for _, crd := range crds.Items {
		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  crd.Spec.Version,
			Resource: crd.Spec.Names.Plural,
		}

		customResourceList, err := cmd.K8s.Dynamic().Resource(gvr).Namespace(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
		if err != nil && !apierr.IsNotFound(err) {
			return err
		}

		if customResourceList == nil {
			continue
		}

		for i := range customResourceList.Items {
			cr := customResourceList.Items[i]
			retryErr := k8sRetry.RetryOnConflict(k8sRetry.DefaultRetry, func() error { return cmd.removeCustomResourceFinalizers(gvr, cr) })
			if retryErr != nil {
				return fmt.Errorf("deleting finalizer failed: %v", retryErr)
			}
		}
	}

	return nil
}

func (cmd *command) removeCustomResourceFinalizers(customResource schema.GroupVersionResource, cr2 unstructured.Unstructured) error {
	// Retrieve the latest version of Custom Resource before attempting update
	// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
	res, err := cmd.K8s.Dynamic().Resource(customResource).Namespace(cr2.GetNamespace()).Get(context.Background(), cr2.GetName(), metav1.GetOptions{})
	if err != nil && !apierr.IsNotFound(err) {
		return err
	}

	if res == nil {
		return nil
	}

	if len(res.GetFinalizers()) > 0 {
		fmt.Printf("Deleting finalizer for \"%s\" %s\n", res.GetName(), cr2.GetKind())
		res.SetFinalizers(nil)
		_, err := cmd.K8s.Dynamic().Resource(customResource).Namespace(res.GetNamespace()).Update(context.Background(), res, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("Deleted finalizer for \"%s\" %s\n", res.GetName(), res.GetKind())
	}

	if !cmd.opts.KeepCRDs {
		err = cmd.K8s.Dynamic().Resource(customResource).Namespace(res.GetNamespace()).Delete(context.Background(), res.GetName(), metav1.DeleteOptions{})
		if err != nil && !apierr.IsNotFound(err) {
			return err
		}
	}
	return nil
}
func (cmd *command) deleteKymaNamespaces() error {

	namespaces := [3]string{"istio-system", "kyma-system", "kyma-integration"}

	var wg sync.WaitGroup
	wg.Add(len(namespaces))
	finishedCh := make(chan bool)
	errorCh := make(chan error)

	for _, namespace := range namespaces {
		go func(ns string) {
			defer wg.Done()
			err := retry.Do(func() error {

				if err := cmd.K8s.Static().CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
					errorCh <- err
				}

				nsT, err := cmd.K8s.Static().CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
				if err != nil && !apierr.IsNotFound(err) {
					errorCh <- err
				} else if apierr.IsNotFound(err) {
					return nil
				}

				return errors.Wrapf(err, "\"%s\" Namespace still exists in \"%s\" Phase", nsT.Name, nsT.Status.Phase)
			})
			if err != nil {
				errorCh <- err
				return
			}
			fmt.Printf("\"%s\" Namespace is removed\n", ns)
		}(namespace)
	}

	go func() {
		wg.Wait()
		close(errorCh)
		close(finishedCh)
	}()

	// process deletion results
	var errWrapped error
	for {
		select {
		case <-finishedCh:
			return errWrapped
		case err := <-errorCh:
			if err != nil {
				if errWrapped == nil {
					errWrapped = err
				} else {
					errWrapped = errors.Wrap(err, errWrapped.Error())
				}
			}
		}
	}
}

func (cmd *command) deleteKymaCRDs() error {
	err := cmd.deleteCRDsByLabelWithRetry("reconciler.kyma-project.io/managed-by=reconciler")
	if err != nil {
		return errors.Wrapf(err, "Failed to delete resource")
	}

	err = cmd.deleteCRDsByLabelWithRetry("install.operator.istio.io/owning-resource-namespace=istio-system")
	if err != nil {
		return errors.Wrapf(err, "Failed to delete resource")
	}

	fmt.Printf("Kyma CRDs successfully uninstalled")

	return nil
}

func (cmd *command) deleteCRDsByLabelWithRetry(labelSelector string) error {
	return retry.Do(func() error {
		if err := cmd.K8s.Dynamic().Resource(crdGvr).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: labelSelector}); err != nil {
			return errors.Wrapf(err, "Error occurred during resources delete: %s", err.Error())
		}
		return nil
	})
}
