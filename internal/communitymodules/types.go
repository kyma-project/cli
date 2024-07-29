package communitymodules

import "k8s.io/apimachinery/pkg/runtime/schema"

// This structure contains only the fields currently in use.
type Modules []Module

type Module struct {
	Name             string    `json:"name,omitempty"`
	Versions         []Version `json:"versions,omitempty"`
	Repository       string    `json:"repository,omitempty"`
	ManagedResources []string  `json:"managedResources,omitempty"`
}

type Version struct {
	Version     string `json:"version,omitempty"`
	ManagerPath string `json:"managerPath,omitempty"`
	Repository  string `json:"repository,omitempty"`
	//CrPath string `json:"crPath,omitempty"`
	DeploymentYaml string `json:"deploymentYaml,omitempty"`
	CrYaml         string `json:"crYaml,omitempty"`
}

var (
	GVRKyma = schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta2",
		Resource: "kymas",
	}
)
