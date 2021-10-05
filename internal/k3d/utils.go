// Package k3d contains special logic to manage installation in k3d clusters
package k3d

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/avast/retry-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// IsK3dCluster checks if the cluster accessible via the kubeclient is a k3d cluster.
// Should this not be possible to determine, an error will be returned.
func IsK3dCluster(kubeClient kubernetes.Interface) (isK3d bool, err error) {

	retryOptions := []retry.Option{
		retry.Delay(2 * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	err = retry.Do(func() error {
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, node := range nodeList.Items {
			if strings.HasPrefix(node.GetName(), "k3d-") {
				isK3d = true
				return nil
			}
		}

		return nil
	}, retryOptions...)
	if err != nil {
		return isK3d, err
	}

	return isK3d, nil
}

// ClusterName finds out the name of the cluster accessible via the kubeclient if it is a k3d cluster.
func ClusterName(kubeClient kubernetes.Interface) (k3dName string, err error) {
	retryOptions := []retry.Option{
		retry.Delay(2 * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	err = retry.Do(func() error {
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{"node-role.kubernetes.io/master": "true"},
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.Background(), listOptions)
		if err != nil {
			return err
		}

		for _, node := range nodeList.Items {
			nodeName := node.GetName()
			if !strings.HasPrefix(nodeName, "k3d-") {
				k3dName = ""
				return errors.New("cluster is not a k3d cluster")
			}
			// K3d cluster name can be derived from master node names, which has the form k3d-<cluster-name>-server-<id>.
			// E.g., with the Kyma CLI default flags k3d-kyma-server-0
			k3dName = strings.TrimSuffix(strings.TrimRightFunc(strings.TrimPrefix(nodeName, "k3d-"), func(r rune) bool {
				return unicode.IsNumber(r) || r == '-'
			}), "-server")
		}

		return nil
	}, retryOptions...)
	if err != nil {
		return k3dName, err
	}

	return k3dName, nil
}
