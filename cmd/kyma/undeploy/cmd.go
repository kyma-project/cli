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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"strings"
	"sync"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/spf13/cobra"
	k8sRetry "k8s.io/client-go/util/retry"
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

	err = cmd.undeployKyma()
	if err != nil {
		return err
	}

	fmt.Println("Kyma successfully removed.")

	return nil
}

func (cmd *command) undeployKyma() error {
	err := cmd.deleteKymaNamespaces()
	if err != nil {
		return err
	}

	if !cmd.opts.KeepCRDs {
		return cmd.deleteKymaCrds()
	}
	return nil
}

func (cmd *command) deleteKymaNamespaces() error {
	err := cmd.cleanupFinalizers()
	if err != nil {
		return err
	}

	namespaces := [3]string{"istio-system", "kyma-system", "kyma-integration"}

	var wg sync.WaitGroup
	wg.Add(len(namespaces))
	finishedCh := make(chan bool)
	errorCh := make(chan error)
	// start deletion in goroutines
	for _, namespace := range namespaces {
		ns2 := namespace
		err := retry.Do(func() error {
			// Check if there are any running Pods left on the namespace
			pods, err := cmd.K8s.Static().CoreV1().Pods(ns2).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				errorCh <- err
			}

			if len(pods.Items) > 0 {
				for _, pod := range pods.Items {
					if pod.Status.Phase == v1.PodRunning {
						return errors.New(fmt.Sprintf("\"%s\" Namespace could not be deleted because of the running \"%s\" Pod. Trying again..", ns2, pod.Name))
					}
				}
			}
			return nil
		})

		if err != nil {
			fmt.Printf("\"%s\" Namespace could not be deleted because of running Pod(s)", ns2)
			wg.Done()
			continue
		}

		go func(ns string) {
			defer wg.Done()
			err = retry.Do(func() error {
				//remove namespace
				if err := cmd.K8s.Static().CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
					errorCh <- err
				}

				nsT, err := cmd.K8s.Static().CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
				if err != nil && !apierr.IsNotFound(err) {
					errorCh <- err
				} else if apierr.IsNotFound(err) {
					return nil
				}
				fmt.Printf("\"%s\" Namespace still exists in \"%s\" Phase", nsT.Name, nsT.Status.Phase)
				return fmt.Errorf("\"%s\" Namespace still exists in \"%s\" Phase", nsT.Name, nsT.Status.Phase)
			})
			if err != nil {
				errorCh <- err
			}
			fmt.Printf("\"%s\" Namespace is removed", ns)
		}(namespace)
	}

	// wait until parallel deletion is finished
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

func (cmd *command) cleanupFinalizers() error {
	// All the hacks below should be deleted after this issue is done: https://github.com/kyma-project/kyma/issues/11298
	//HACK: Delete finalizers of leftover Secret
	secrets, err := cmd.K8s.Static().CoreV1().Secrets(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{LabelSelector: "serverless.kyma-project.io/config=credentials"})
	if err != nil && !apierr.IsNotFound(err) {
		return err
	}
	if secrets != nil {
		for _, secret := range secrets.Items {
			secret := secret
			if len(secret.GetFinalizers()) > 0 {
				secret.SetFinalizers(nil)
				if _, err := cmd.K8s.Static().CoreV1().Secrets(secret.GetNamespace()).Update(context.Background(), &secret, metav1.UpdateOptions{}); err != nil {
					return err
				}
				fmt.Printf("Deleted finalizer from \"%s\" Secret", secret.GetName())
			}
		}
	}
	//HACK: Delete finalizers of leftover Custom Resources
	selector, err := cmd.prepareKymaCrdLabelSelector()
	if err != nil {
		return err
	}

	crds, err := cmd.apixClient.CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil && !apierr.IsNotFound(err) {
		return err
	}

	if crds != nil {
		for _, crd := range crds.Items {
			customResource := schema.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  crd.Spec.Version,
				Resource: crd.Spec.Names.Plural,
			}

			customResourceList, err := cmd.K8s.Dynamic().Resource(customResource).Namespace(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
			if err != nil && !apierr.IsNotFound(err) {
				return err
			}
			if customResourceList != nil {
				for _, cr := range customResourceList.Items {
					cr2 := cr
					retryErr := k8sRetry.RetryOnConflict(k8sRetry.DefaultRetry, func() error {
						// Retrieve the latest version of Custom Resource before attempting update
						// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
						res, err := cmd.K8s.Dynamic().Resource(customResource).Namespace(cr2.GetNamespace()).Get(context.Background(), cr2.GetName(), metav1.GetOptions{})
						if err != nil && !apierr.IsNotFound(err) {
							return err
						}
						if res != nil {
							if len(res.GetFinalizers()) > 0 {
								fmt.Printf("Deleting finalizer for \"%s\" %s", res.GetName(), cr2.GetKind())
								res.SetFinalizers(nil)
								_, err := cmd.K8s.Dynamic().Resource(customResource).Namespace(res.GetNamespace()).Update(context.Background(), res, metav1.UpdateOptions{})
								if err != nil {
									return err
								}
								fmt.Printf("Deleted finalizer for \"%s\" %s", res.GetName(), res.GetKind())
							}
							if !cmd.opts.KeepCRDs {
								err = cmd.K8s.Dynamic().Resource(customResource).Namespace(res.GetNamespace()).Delete(context.Background(), res.GetName(), metav1.DeleteOptions{})
								if err != nil && !apierr.IsNotFound(err) {
									return err
								}
							}
						}
						return nil
					})
					if retryErr != nil {
						return fmt.Errorf("deleting finalizer failed: %v", retryErr)
					}
				}
			}
		}
	}
	return nil
}

func (cmd *command) deleteKymaCrds() error {
	fmt.Printf("Uninstalling CRDs labeled with: %s=%s", "origin", "kyma")

	selector, err := cmd.prepareKymaCrdLabelSelector()
	if err != nil {
		return err
	}

	gvks := cmd.retrieveKymaCrdGvks()
	for _, gvk := range gvks {
		fmt.Printf("Uninstalling CRDs that belong to apiVersion: %s/%s", gvk.Group, gvk.Version)
		err = cmd.deleteCollectionOfResources(gvk, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return errors.Wrapf(err, "Failed to delete resource")
		}
	}

	fmt.Printf("Kyma CRDs successfully uninstalled")

	return nil
}

func (cmd *command) prepareKymaCrdLabelSelector() (selector labels.Selector, err error) {
	kymaCrdReq, err := labels.NewRequirement("origin", selection.Equals, []string{"kyma"})
	if err != nil {
		return nil, errors.Wrap(err, "Error occurred when preparing Kyma CRD label selector")
	}

	selector = labels.NewSelector()
	selector = selector.Add(*kymaCrdReq)
	return selector, nil
}

func (cmd *command) retrieveKymaCrdGvks() []schema.GroupVersionKind {
	crdGvkV1Beta1 := cmd.crdGvkWith("v1beta1")
	crdGvkV1 := cmd.crdGvkWith("v1")
	return []schema.GroupVersionKind{crdGvkV1Beta1, crdGvkV1}
}

func (cmd *command) deleteCollectionOfResources(gvk schema.GroupVersionKind, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var err error
	err = retry.Do(func() error {
		if err = cmd.K8s.Dynamic().Resource(retrieveGvrFrom(gvk)).DeleteCollection(context.TODO(), opts, listOpts); err != nil {
			return errors.Wrapf(err, "Error occurred during resources delete: %s", err.Error())
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (cmd *command) crdGvkWith(version string) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: version,
		Kind:    "customresourcedefinition",
	}
}

func retrieveGvrFrom(gvk schema.GroupVersionKind) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: pluralForm(gvk.Kind),
	}
}

func pluralForm(singular string) string {
	return fmt.Sprintf("%ss", strings.ToLower(singular))
}
