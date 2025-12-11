package fake

import "github.com/kyma-project/cli.v3/internal/modulesv2/entities"

type CommunityParams struct {
	TemplateName string
	ModuleName   string
	Version      string
	Namespace    string
	SourceURL    string
	Resources    map[string]string
}

func CommunityModuleTemplate(params *CommunityParams) *entities.CommunityModuleTemplate {
	defaults := defaultCommunityParams()

	if params == nil {
		params = defaults
	}

	base := entities.MapBaseModuleTemplateFromParams(
		firstNonEmpty(params.TemplateName, defaults.TemplateName),
		firstNonEmpty(params.ModuleName, defaults.ModuleName),
		firstNonEmpty(params.Version, defaults.Version),
		firstNonEmpty(params.Namespace, defaults.Namespace),
	)

	resources := params.Resources
	if resources == nil {
		resources = defaults.Resources
	}

	return entities.NewCommunityModuleTemplate(
		base,
		firstNonEmpty(params.SourceURL, defaults.SourceURL),
		resources,
	)
}

func defaultCommunityParams() *CommunityParams {
	return &CommunityParams{
		TemplateName: "sample-community-template-1.0.0",
		ModuleName:   "sample-community-template",
		Version:      "1.0.0",
		Namespace:    "",
		SourceURL:    "",
		Resources:    map[string]string{},
	}
}
