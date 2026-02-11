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
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modulesv2/repository"
	"github.com/kyma-project/cli.v3/internal/out"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

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

const (
	lifecycleManagerRepo             = "kyma-project/lifecycle-manager"
	lifecycleManagerLatestReleaseURL = "https://api.github.com/repos/" + lifecycleManagerRepo + "/releases/latest"
	moduleTemplateCRDPathTemplate    = "https://raw.githubusercontent.com/" + lifecycleManagerRepo + "/refs/tags/%s/config/crd/bases/operator.kyma-project.io_moduletemplates.yaml"
)

// CRDEnsurer ensures the ModuleTemplate CRD is installed and up-to-date.
type CRDEnsurer struct {
	client                    kube.Client
	clusterMetadataRepository repository.ClusterMetadataRepository
	httpClient                *http.Client
	remoteURL                 string
}

// NewCRDEnsurer creates a new CRDEnsurer instance.
func NewCRDEnsurer(client kube.Client, clusterMetadataRepository repository.ClusterMetadataRepository, httpClient *http.Client, crdURL string) *CRDEnsurer {
	return &CRDEnsurer{
		client:                    client,
		clusterMetadataRepository: clusterMetadataRepository,
		httpClient:                httpClient,
		remoteURL:                 crdURL,
	}
}

func newCRDEnsurer(client kube.Client) *CRDEnsurer {
	return NewCRDEnsurer(client, repository.NewClusterMetadataRepository(client), http.DefaultClient, "")
}

// EnsureCRD ensures the ModuleTemplate CRD is installed and up-to-date on the cluster.
// If the CRD is missing or outdated, it prompts the user (unless force is true) and applies it.
func EnsureCRD(kymaConfig *cmdcommon.KymaConfig, force bool) clierror.Error {
	kubeClient, clierr := kymaConfig.KubeClientConfig.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	ensurer := newCRDEnsurer(kubeClient)
	if err := ensurer.run(kymaConfig.Ctx, force); err != nil {
		return clierror.Wrap(err, clierror.New("failed to install required dependencies"))
	}

	return nil
}

func (e *CRDEnsurer) run(ctx context.Context, force bool) error {
	if e.isKLMManaged(ctx) {
		return nil
	}

	storedCRD, err := e.fetchStoredCRD(ctx)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to access stored CRD: %w", err)
	}

	crdNotFound := errors.IsNotFound(err) || storedCRD == nil

	remoteCRD, err := e.fetchRemoteCRD(ctx)
	if err != nil {
		out.Debugfln("failed to fetch CRD from remote: %w", err)
		return nil
	}

	equal, err := e.specEqual(storedCRD, remoteCRD)
	if err != nil {
		return err
	}
	if equal {
		return nil
	}

	if !force {
		if crdNotFound {
			if err := e.promptForCRDInstallation(); err != nil {
				return err
			}
		} else {
			// CRD exists but is outdated - update is optional
			if !e.promptForCRDUpdate() {
				// User declined update, continue with existing CRD
				return nil
			}
		}
	}

	return e.applyCRD(ctx, remoteCRD)
}

func (e *CRDEnsurer) promptForCRDInstallation() error {
	out.Msgfln("The ModuleTemplate Custom Resource Definition (CRD) is not installed on this cluster.")
	out.Msgfln("This CRD is required to pull and manage community modules.")
	out.Msgfln("")
	out.Msgfln("The CLI will download and install the CRD from the latest release of:")
	out.Msgfln("  https://github.com/%s", lifecycleManagerRepo)
	out.Msgfln("")
	out.Msgfln("Tip: You can use the --force flag to automatically approve this installation.")
	out.Msgfln("")

	confirmPrompt := prompt.NewBool("Do you want to proceed with the installation?", false)
	confirmed, err := confirmPrompt.Prompt()
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}
	if !confirmed {
		return fmt.Errorf("installation cancelled by user")
	}

	return nil
}

func (e *CRDEnsurer) promptForCRDUpdate() bool {
	out.Msgfln("The ModuleTemplate Custom Resource Definition (CRD) on this cluster is not the latest version.")
	out.Msgfln("An updated version is available and recommended for managing community modules.")
	out.Msgfln("")
	out.Msgfln("The CLI will download and apply the updated CRD from the latest release of:")
	out.Msgfln("  https://github.com/%s", lifecycleManagerRepo)
	out.Msgfln("")
	out.Msgfln("Tip: You can use the --force flag to automatically approve this update.")
	out.Msgfln("")

	confirmPrompt := prompt.NewBool("Do you want to proceed with the update?", false)
	confirmed, err := confirmPrompt.Prompt()
	if err != nil {
		out.Errfln("failed to get user confirmation: %v", err)
		return false
	}
	return confirmed
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
	specMap, ok := spec.(map[string]any)
	if !ok {
		return "", fmt.Errorf("CRD spec is not a map")
	}
	normalizedSpec := normalizeSpec(specMap)
	b, err := json.Marshal(normalizedSpec)
	if err != nil {
		return "", fmt.Errorf("marshal spec: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func (e *CRDEnsurer) fetchRemoteCRD(ctx context.Context) (*unstructured.Unstructured, error) {
	crdURL := e.remoteURL
	// If remoteURL is empty, fetch the latest release tag and construct the URL
	if crdURL == "" {
		latestTag, err := e.fetchLatestReleaseTag(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest release tag: %w", err)
		}
		crdURL = fmt.Sprintf(moduleTemplateCRDPathTemplate, latestTag)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, crdURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.httpClient.Do(req)
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

// fetchLatestReleaseTag fetches the latest release tag from GitHub API
func (e *CRDEnsurer) fetchLatestReleaseTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lifecycleManagerLatestReleaseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch latest release: status %d: %s", resp.StatusCode, string(body))
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name found in latest release response")
	}

	return release.TagName, nil
}

func (e *CRDEnsurer) isKLMManaged(ctx context.Context) bool {
	return e.clusterMetadataRepository.Get(ctx).IsManagedByKLM
}

func (e *CRDEnsurer) fetchStoredCRD(ctx context.Context) (*unstructured.Unstructured, error) {
	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion(moduleTemplateCRDMeta.APIVersion)
	crd.SetKind(moduleTemplateCRDMeta.Kind)
	crd.SetName(moduleTemplateCRDMeta.Name)

	return e.client.RootlessDynamic().Get(ctx, crd)
}

// specEqual compares the .spec section of two CRDs via SHA-256 digest
func (e *CRDEnsurer) specEqual(a, b *unstructured.Unstructured) (bool, error) {
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

// normalizeSpec removes fields from the CRD spec that are added by Kubernetes
// as server-side defaults, so that comparison between remote and stored CRDs
// is accurate.
func normalizeSpec(spec map[string]any) map[string]any {
	normalized := make(map[string]any)
	for k, v := range spec {
		normalized[k] = v
	}

	// Remove "conversion" field if it's the default value (strategy: None)
	// Kubernetes automatically adds this when a CRD is applied without it
	if conversion, ok := normalized["conversion"].(map[string]any); ok {
		if strategy, ok := conversion["strategy"].(string); ok && strategy == "None" && len(conversion) == 1 {
			delete(normalized, "conversion")
		}
	}

	return normalized
}

func (e *CRDEnsurer) applyCRD(ctx context.Context, remote *unstructured.Unstructured) error {
	if err := e.client.RootlessDynamic().Apply(ctx, remote, false); err != nil {
		return fmt.Errorf("failed to apply ModuleTemplate CRD on the target cluster: %w", err)
	}
	return nil
}
