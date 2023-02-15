package deploy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli/internal/kube"
)

const (
	certManagerURL = "https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml"
)

// CertManager deploys the Kyma CR. If no kymaCRPath is provided, it deploys the default CR.
func CertManager(ctx context.Context, k8s kube.KymaKube, certManagerVersion string, dryRun bool) error {
	result := bytes.Buffer{}

	// Get the data
	resp, err := http.Get(fmt.Sprintf(certManagerURL, certManagerVersion))
	if err != nil {
		return fmt.Errorf("could not download cert-manager: %w", err)
	}

	certManagerBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not write cert-manager data to yaml: %w", err)
	}
	result.Write(certManagerBytes)

	if dryRun {
		fmt.Printf("%s---\n", result.String())
		return nil
	}

	return k8s.Apply(ctx, result.Bytes())
}
