package precheck

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

type InstallCRD struct {
	client                    kube.Client
	clusterMetadataRepository repository.ClusterMetadataRepository
	httpClient                *http.Client
	remoteURL                 string
}

type crdMeta struct {
	Name       string
	APIVersion string
	Kind       string
}

var moduleTemplateCRDMeta = crdMeta{
	Name:       "moduletemplates.operator.kyma-project.io",
	APIVersion: "apiextensions.k8s.io/v1",
	Kind:       "CustomResourceDefinition",
}

const moduleTemplateCRDURL = "https://raw.githubusercontent.com/kyma-project/lifecycle-manager/refs/heads/main/config/crd/bases/operator.kyma-project.io_moduletemplates.yaml"

func NewInstallCRD(client kube.Client, clusterMetadataRepository repository.ClusterMetadataRepository, httpClient *http.Client, crdURL string) *InstallCRD {
	return &InstallCRD{
		client:                    client,
		clusterMetadataRepository: clusterMetadataRepository,
		httpClient:                httpClient,
		remoteURL:                 crdURL,
	}
}

func newPreCheck(client kube.Client) *InstallCRD {
	return NewInstallCRD(client, repository.NewClusterMetadataRepository(client), http.DefaultClient, moduleTemplateCRDURL)
}

// RunInstallCRD bootstraps dependencies and runs the InstallCRD; returns clierror.Error for easy reuse.
func RunInstallCRD(kymaConfig *cmdcommon.KymaConfig) clierror.Error {
	kubeClient, clierr := kymaConfig.KubeClientConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	preCheck := newPreCheck(kubeClient)
	if err := preCheck.run(kymaConfig.Ctx); err != nil {
		return clierror.Wrap(err, clierror.New("failed to run pre-checks"))
	}

	return nil
}

func (p *InstallCRD) run(ctx context.Context) error {
	if p.isKLMManaged(ctx) {
		return nil
	}

	storedCRD, err := p.fetchStoredCRD(ctx)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to access stored CRD: %w", err)
	}

	remoteCRD, err := p.fetchRemoteCRD(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch CRD from remote: %w", err)
	}

	equal, err := p.specEqual(storedCRD, remoteCRD)
	if err != nil {
		return err
	}
	if equal {
		return nil
	}

	return p.applyCRD(ctx, remoteCRD)
}

// crdSpecDigest returns a SHA-256 digest of the CRD's .spec section
func crdSpecDigest(u *unstructured.Unstructured) (string, error) {
	if u == nil || u.Object == nil {
		return "", nil
	}
	spec, ok := u.Object["spec"]
	if !ok {
		return "", fmt.Errorf("CRD missing spec")
	}
	b, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("marshal spec: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func (p *InstallCRD) fetchRemoteCRD(ctx context.Context) (*unstructured.Unstructured, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.remoteURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch CRD: status %d: %s", resp.StatusCode, string(body))
	}

	yamlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert YAML to JSON, then to Unstructured
	jsonBytes, err := k8syaml.ToJSON(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CRD YAML to JSON: %w", err)
	}

	var obj map[string]any
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CRD JSON: %w", err)
	}

	return &unstructured.Unstructured{Object: obj}, nil
}

func (p *InstallCRD) isKLMManaged(ctx context.Context) bool {
	return p.clusterMetadataRepository.Get(ctx).IsManagedByKLM
}

func (p *InstallCRD) fetchStoredCRD(ctx context.Context) (*unstructured.Unstructured, error) {
	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion(moduleTemplateCRDMeta.APIVersion)
	crd.SetKind(moduleTemplateCRDMeta.Kind)
	crd.SetName(moduleTemplateCRDMeta.Name)

	return p.client.RootlessDynamic().Get(ctx, crd)
}

// specEqual compares the .spec section of two CRDs via SHA-256 digest
func (p *InstallCRD) specEqual(a, b *unstructured.Unstructured) (bool, error) {
	var ad, bd string
	var err error

	if a != nil && a.Object != nil {
		ad, err = crdSpecDigest(a)
		if err != nil {
			return false, fmt.Errorf("failed to compute digest of stored CRD: %w", err)
		}
	}

	bd, err = crdSpecDigest(b)
	if err != nil {
		return false, fmt.Errorf("failed to compute digest of remote CRD: %w", err)
	}

	return ad == bd, nil
}

func (p *InstallCRD) applyCRD(ctx context.Context, remote *unstructured.Unstructured) error {
	if err := p.client.RootlessDynamic().Apply(ctx, remote, false); err != nil {
		return fmt.Errorf("failed to apply ModuleTemplate CRD on the target cluster: %w", err)
	}
	return nil
}
