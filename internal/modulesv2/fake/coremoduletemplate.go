package fake

import "github.com/kyma-project/cli.v3/internal/modulesv2/entities"

type Params struct {
	TemplateName string
	ModuleName   string
	Version      string
	Channel      string
	Namespace    string
}

func CoreModuleTemplate(params *Params) *entities.CoreModuleTemplate {
	defaults := defaultParams()

	if params == nil {
		params = defaults
	}

	return entities.NewCoreModuleTemplateFromParams(
		firstNonEmpty(params.TemplateName, defaults.TemplateName),
		firstNonEmpty(params.ModuleName, defaults.ModuleName),
		firstNonEmpty(params.Version, defaults.Version),
		firstNonEmpty(params.Channel, defaults.Channel),
		firstNonEmpty(params.Namespace, defaults.Namespace),
	)
}

func defaultParams() *Params {
	return &Params{
		TemplateName: "sample-template-0.0.1",
		ModuleName:   "sample-template",
		Version:      "0.0.1",
		Channel:      "fast",
		Namespace:    "kyma-system",
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
