package version

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
)

type KymaVersion struct {
	semanticVersion *semver.Version
	stringVersion   string
}

type UpgradeScenario string

const (
	UpgradeEqualVersion UpgradeScenario = "upgrading to same version"
	UpgradePossible     UpgradeScenario = "upgrade is possible"
	UpgradeUndetermined UpgradeScenario = "upgrade path cannot be determined"
)

func NewKymaVersion(kymaVersion string) (KymaVersion, error) {
	re := regexp.MustCompile(`^[1-9]\.`)
	res := re.FindString(kymaVersion)
	if res != "" {
		v, err := semver.NewVersion(kymaVersion)
		if err != nil {
			return KymaVersion{semanticVersion: v}, errors.Wrapf(err, "Version is not a semver: %s", kymaVersion)
		}
		return KymaVersion{semanticVersion: v, stringVersion: kymaVersion}, nil
	}

	return KymaVersion{stringVersion: kymaVersion}, nil
}

func NewNoVersion() KymaVersion {
	return KymaVersion{stringVersion: "N/A"}
}

func (kv *KymaVersion) IsReleasedVersion() bool {
	return kv.semanticVersion != nil
}

func (kv *KymaVersion) IsCompatibleWith(upgradeVersion KymaVersion) UpgradeScenario {
	if kv.stringVersion == upgradeVersion.stringVersion {
		return UpgradeEqualVersion
	}
	if err := checkCompatibility(kv.stringVersion, upgradeVersion.stringVersion); err != nil {
		return UpgradeUndetermined
	}
	return UpgradePossible
}

func (kv *KymaVersion) IsKyma1() bool {
	return kv.semanticVersion.Major() == 1
}

func (kv *KymaVersion) IsKyma2() bool {
	return kv.semanticVersion.Major() == 2
}

func (kv *KymaVersion) None() bool {
	return kv.stringVersion == "N/A"
}

func (kv *KymaVersion) String() string {
	return kv.stringVersion
}

//GetCurrentKymaVersion determines the semanticVersion of kyma installed in the cluster via the provided kubernetes client
func GetCurrentKymaVersion(k8s kube.KymaKube) (KymaVersion, error) {
	isKyma2, err := checkKyma2(k8s)
	if err != nil {
		return KymaVersion{}, err
	}

	if isKyma2 {
		return getKyma2Version(k8s)
	}
	return getKyma1Version(k8s)
}

func checkKyma2(k8s kube.KymaKube) (bool, error) {
	deps, err := getDeployments(k8s)
	if err != nil {
		return false, err
	}
	if len(deps.Items) == 0 {
		return false, nil
	}
	return true, nil
}

func getDeployments(k8s kube.KymaKube) (*v1.DeploymentList, error) {
	return k8s.Static().AppsV1().Deployments("kyma-system").List(context.Background(), metav1.ListOptions{LabelSelector: "reconciler.kyma-project.io/managed-by=reconciler"})
}

func getKyma2Version(k8s kube.KymaKube) (KymaVersion, error) {
	deployments, err := getDeployments(k8s)
	if err != nil {
		return KymaVersion{}, err
	}
	if len(deployments.Items) == 0 {
		return NewNoVersion(), nil
	}
	version, err := NewKymaVersion(deployments.Items[0].Labels["reconciler.kyma-project.io/origin-version"])
	if err != nil {
		return version, err
	}

	return version, nil
}

func getKyma1Version(k8s kube.KymaKube) (KymaVersion, error) {
	pods, err := k8s.Static().CoreV1().Pods("kyma-installer").List(context.Background(), metav1.ListOptions{LabelSelector: "name=kyma-installer"})
	if err != nil {
		return KymaVersion{}, err
	}

	if len(pods.Items) == 0 {
		return NewNoVersion(), nil
	}

	imageParts := strings.Split(pods.Items[0].Spec.Containers[0].Image, ":")
	if len(imageParts) < 2 {
		return NewNoVersion(), nil
	}

	version, err := NewKymaVersion(imageParts[1])
	if err != nil {
		return version, err
	}

	return version, nil
}
