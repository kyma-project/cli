package clusterinfo

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"

	"github.com/avast/retry-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func isK3dCluster(ctx context.Context, kubeClient kubernetes.Interface) (isK3d bool, err error) {

	retryOptions := []retry.Option{
		retry.Delay(2 * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	err = retry.Do(func() error {
		nodeList, err := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
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

func k3dClusterName(ctx context.Context, kubeClient kubernetes.Interface) (k3dName string, err error) {
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
		nodeList, err := kubeClient.CoreV1().Nodes().List(ctx, listOptions)
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
