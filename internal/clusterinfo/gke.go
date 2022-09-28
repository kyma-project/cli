package clusterinfo

import (
	"context"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func isGkeCluster(ctx context.Context, kubeClient kubernetes.Interface) (isGke bool, err error) {
	matcher, err := regexp.Compile(`^v\d+\.\d+\.\d+-gke\.\d+$`)
	if err != nil {
		return false, err
	}

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
			match := matcher.MatchString(node.Status.NodeInfo.KubeProxyVersion)
			if match {
				isGke = true
				return nil
			}
		}

		return nil
	}, retryOptions...)
	if err != nil {
		return isGke, err
	}

	return isGke, nil
}
