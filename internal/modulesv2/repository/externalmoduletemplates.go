package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
)

type ExternalModuleTemplateRepository interface {
	Get(urls []string) ([]kyma.ModuleTemplate, error)
}

type externalModuleTemplateRepository struct{}

func NewExternalModuleTemplateRepository() *externalModuleTemplateRepository {
	return &externalModuleTemplateRepository{}
}

func (r *externalModuleTemplateRepository) Get(urls []string) ([]kyma.ModuleTemplate, error) {
	externalModules := []kyma.ModuleTemplate{}

	for _, url := range urls {
		externalModuleTemplatesList, err := getFileFromURL(url)
		if err != nil {
			return nil, fmt.Errorf("failed to get community modules definitions: %v", err)
		}

		var result []kyma.ModuleTemplate
		if err := json.Unmarshal(externalModuleTemplatesList, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal module template: %w", err)
		}

		externalModules = append(externalModules, result...)
	}

	return externalModules, nil
}

func getFileFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download resource from %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource body: %w", err)
	}

	return body, nil
}
