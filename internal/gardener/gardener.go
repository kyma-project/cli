// Package Gardener contains special logic to manage installation in Gardener clusters
package gardener

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Domain fetches the domain for the gardener cluster accessible via the kubeclient
func Domain(kubeClient kubernetes.Interface) (domainName string, err error) {
	err = retry.Do(func() error {
		configMap, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "shoot-info", metav1.GetOptions{})

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
	}, retry.Delay(2*time.Second), retry.Attempts(3), retry.DelayType(retry.FixedDelay))

	if err != nil {
		return "", err
	}

	return domainName, nil
}
