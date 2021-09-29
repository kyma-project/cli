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
	version *semver.Version
	stringVersion string
}

type UpgradeScenario string
const (
	UpgradeEqualVersion UpgradeScenario = "upgrading to same version"
	UpgradablePossible UpgradeScenario = "upgrade is possible"
	UpgradeUndetermined UpgradeScenario = "upgrade path cannot be determined"

)

func NewKymaVersion(kymaVersion string) (KymaVersion, error){
	re := regexp.MustCompile(`^[1-9]`)
	res := re.FindString(kymaVersion)
	if res != "" {
		v, err := semver.NewVersion(kymaVersion)
		if err != nil {
			return KymaVersion{version:  v}, errors.Wrapf(err, "Version is not a semver: %s", kymaVersion)
		}
		return KymaVersion{version: v, stringVersion: kymaVersion}, nil
	}

	return KymaVersion{stringVersion:  kymaVersion}, nil

}

func (kv *KymaVersion) IsReleasedVersion () bool{
	return kv.version != nil
}

func (kv *KymaVersion) IsCompatibleVersion(upgradeVersion KymaVersion) UpgradeScenario {
	if kv.stringVersion == upgradeVersion.stringVersion {
		return UpgradeEqualVersion
	}
	if err := checkCompatibility(kv.stringVersion, upgradeVersion.stringVersion); err != nil {
		return UpgradeUndetermined
	}
	return UpgradablePossible

}

func (kv *KymaVersion) IsKyma1() bool{
	return kv.version.Major() == 1
}

func (kv *KymaVersion) IsKyma2() bool{
	return kv.version.Major() == 2
}

func (kv *KymaVersion) HasNoVersion() bool {
	return kv.stringVersion == "N/A"
}

func (kv *KymaVersion) String() string{
	return kv.stringVersion
}

//KymaVersion determines the version of kyma installed in the cluster via the provided kubernetes client
func GetCurrentKymaVersion(k8s kube.KymaKube) (KymaVersion, error) {
	isKyma2, err := checkKyma2(k8s)
	var version string
	if err != nil {
		return KymaVersion{}, err
	}
	if isKyma2 {
		//Check for kyma 2
		version, err = getKyma2Version(k8s)
	} else {
		version, err =  getKyma1Version(k8s)
	}

	if err != nil {
		return KymaVersion{}, err
	}
	return NewKymaVersion(version)
}


func getDeployments(k8s kube.KymaKube) (*v1.DeploymentList, error) {
	return k8s.Static().AppsV1().Deployments("kyma-system").List(context.Background(), metav1.ListOptions{LabelSelector: "reconciler.kyma-project.io/managed-by=reconciler"})
}

func getKyma2Version(k8s kube.KymaKube) (string, error) {
	deps, err := getDeployments(k8s)
	if err != nil {
		return "N/A", err
	}
	if len(deps.Items) == 0 {
		return "N/A", nil
	}
	return deps.Items[0].Labels["reconciler.kyma-project.io/origin-version"], nil
}

func getKyma1Version(k8s kube.KymaKube) (string, error) {
	pods, err := k8s.Static().CoreV1().Pods("kyma-installer").List(context.Background(), metav1.ListOptions{LabelSelector: "name=kyma-installer"})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "N/A", nil
	}

	imageParts := strings.Split(pods.Items[0].Spec.Containers[0].Image, ":")
	if len(imageParts) < 2 {
		return "N/A", nil
	}

	return imageParts[1], nil
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
