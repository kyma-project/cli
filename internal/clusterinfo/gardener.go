// Package Gardener contains special logic to manage installation in Gardener clusters
package clusterinfo

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getGardenerDomain(ctx context.Context, kubeClient kubernetes.Interface) (domainName string, err error) {
	err = retry.Do(func() error {
		configMap, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(ctx, "shoot-info", metav1.GetOptions{})

		if err != nil {
			if apierr.IsNotFound(err) {
				return nil
			}
			return err
		}

		domainName = configMap.Data["domain"]
		if domainName == "" {
			return fmt.Errorf("domain is empty in %s configmap", "shoot-info")
		}

		return nil
	}, retry.Delay(2*time.Second), retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Context(ctx))
	if err != nil {
		return "", err
	}

	return domainName, nil
}
