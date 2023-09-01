package deploy

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/pkg/errors"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/internal/kustomize"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

const (
	wildCardRoleAndAssignment = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kyma-cli-provisioned-wildcard
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: lifecycle-manager-wildcard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kyma-cli-provisioned-wildcard
subjects:
- kind: ServiceAccount
  name: lifecycle-manager-controller-manager
  namespace: kcp-system`
)

// Bootstrap deploys the kustomization files for the prerequisites for Kyma.
// Returns true if the Kyma CRD was deployed.
func Bootstrap(
	ctx context.Context, kustomizations []string, k8s kube.KymaKube, filters []kio.Filter, addWildCard, force bool,
	dryRun bool, isInKcpMode bool,
) (bool, error) {
	var defs []kustomize.Definition
	// defaults
	for _, k := range kustomizations {
		parsed, err := kustomize.ParseKustomization(k)
		if err != nil {
			return false, err
		}
		defs = append(defs, parsed)
	}

	// build manifests
	manifests, err := kustomize.BuildMany(defs, filters)
	if err != nil {
		return false, err
	}

	if addWildCard {
		manifests = append(manifests, []byte(wildCardRoleAndAssignment)...)
	}

	manifestObjs, err := parseManifests(k8s, manifests, dryRun)
	if err != nil {
		return false, err
	}
	manifestObjs = filterCRD(manifestObjs)
	if isInKcpMode {
		if err = patchDeploymentWithInKcpModeFlag(manifestObjs); err != nil {
			return false, err
		}
	}
	err = applyManifests(
		ctx, k8s, manifests, applyOpts{
			dryRun, force, defaultRetries, defaultInitialBackoff}, manifestObjs,
	)
	if err != nil {
		return false, err
	}

	return hasKyma(string(manifests))
}

// we have to manually configure CreationTimestamp for CRD until this bug get fixed
// https://github.com/kubernetes-sigs/kustomize/issues/5031
func filterCRD(objs []ctrlClient.Object) []ctrlClient.Object {
	var filteredObjs []ctrlClient.Object
	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == "CustomResourceDefinition" {
			obj.SetCreationTimestamp(metav1.Now())
		}
		filteredObjs = append(filteredObjs, obj)
	}
	return filteredObjs
}

func patchDeploymentWithInKcpModeFlag(manifestObjs []ctrlClient.Object) error {
	var deployment *appsv1.Deployment
	for _, manifest := range manifestObjs {
		manifestJSON, err := getManifestJSONForDeployment(manifest)
		if err != nil {
			return err
		}
		if manifestJSON == nil {
			continue
		}
		deployment, err = getDeployment(manifestJSON)
		if err != nil {
			return err
		}
		deploymentArgs := deployment.Spec.Template.Spec.Containers[0].Args
		hasKcpFlag := checkDeploymentHasKcpFlag(deploymentArgs)
		if !hasKcpFlag {
			deployment.Spec.Template.Spec.Containers[0].Args = append(deploymentArgs, "--in-kcp-mode")
			err := patchManifest(deployment, manifest)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

var ErrPatchManifestType = errors.New("failed to cast manifest object to Unstructured")

func patchManifest(deployment *appsv1.Deployment, manifest ctrlClient.Object) error {
	manifestJSON, err := json.Marshal(deployment)
	if err != nil {
		return err
	}
	unstrct, ok := manifest.(*unstructured.Unstructured)
	if !ok {
		return ErrPatchManifestType
	}

	err = json.Unmarshal(manifestJSON, &unstrct.Object)
	if err != nil {
		return err
	}

	return nil
}

func getManifestJSONForDeployment(manifest ctrlClient.Object) ([]byte, error) {
	if manifest.GetObjectKind().GroupVersionKind().Kind != "Deployment" {
		return nil, nil
	}
	if manifestObj, success := manifest.(*unstructured.Unstructured); success {
		var manifestJSON []byte
		var err error
		if manifestJSON, err = json.Marshal(manifestObj.Object); err != nil {
			return nil, err
		}
		return manifestJSON, nil

	}
	return nil, nil
}

func getDeployment(manifestJSON []byte) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if err := json.Unmarshal(manifestJSON, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func checkDeploymentHasKcpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--in-kcp-mode" || arg == "--in-kcp-mode=true" {
			return true
		}
	}

	return false
}

func checkDeploymentReadiness(objs []ctrlClient.Object, k8s kube.KymaKube) error {
	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind != "Deployment" {
			continue
		}
		if err := k8s.WaitDeploymentStatus(
			obj.GetNamespace(), obj.GetName(), appsv1.DeploymentAvailable, corev1.ConditionTrue,
		); err != nil {
			return err
		}
	}
	return nil
}

// hasKyma checks if the given manifest contains the Kyma CRD
func hasKyma(manifest string) (bool, error) {
	r, err := regexp.Compile(`(names:)(?:[.\s\S]*)(kind: Kyma)(?:[.\s\S]*)(plural: kymas)`)
	if err != nil {
		return false, err
	}
	return r.MatchString(manifest), nil
}
