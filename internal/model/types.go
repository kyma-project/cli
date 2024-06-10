package model

type Module []struct {
	Name string `json:"name,omitempty"`
	//Documentation       string              `json:"documentation,omitempty"`
	//Repository          string              `json:"repository,omitempty"`
	//ManagedResources    []string            `json:"managedResources,omitempty"`
	Manageable bool `json:"manageable,omitempty"`
	//LatestGithubRelease LatestGithubRelease `json:"latestGithubRelease,omitempty"`
	Versions []Versions `json:"versions,omitempty"`
}

type Versions struct {
	Version  string   `json:"version,omitempty"`
	Channels []string `json:"channels,omitempty"`
	//Documentation string      `json:"documentation,omitempty"`
	//Repository    string      `json:"repository,omitempty"`
	ManagerPath  string      `json:"managerPath,omitempty"`
	ManagerImage string      `json:"managerImage,omitempty"`
	Resources    []Resources `json:"resources,omitempty"`
}

type Resources struct {
	//APIVersion string   `json:"apiVersion,omitempty"`
	//Kind string `json:"kind,omitempty"`
	//Metadata   Metadata `json:"metadata,omitempty"`
	Spec Spec `json:"spec,omitempty"`
}

type Spec struct {
	Group       string        `json:"group,omitempty"`
	Names       Names         `json:"names,omitempty"`
	Scope       string        `json:"scope,omitempty"`
	ApiVersions []ApiVersions `json:"versions,omitempty"`
}

type Names struct {
	Kind     string `json:"kind,omitempty"`
	ListKind string `json:"listKind,omitempty"`
	Plural   string `json:"plural,omitempty"`
	Singular string `json:"singular,omitempty"`
}

type ApiVersions struct {
	//AdditionalPrinterColumns []AdditionalPrinterColumns `json:"additionalPrinterColumns,omitempty"`
	Name string `json:"name,omitempty"`
	//Schema                   Schema                     `json:"schema,omitempty"`
}

//type LatestGithubRelease struct {
//	Repository     string `json:"repository,omitempty"`
//	DeploymentYaml string `json:"deploymentYaml,omitempty"`
//	CrYaml         string `json:"crYaml,omitempty"`
//}

//type Annotations struct {
//	ControllerGenKubebuilderIoVersion string `json:"controller-gen.kubebuilder.io/version,omitempty"`
//}

//type Labels struct {
//	AppKubernetesIoComponent string `json:"app.kubernetes.io/component,omitempty"`
//	AppKubernetesIoInstance  string `json:"app.kubernetes.io/instance,omitempty"`
//	AppKubernetesIoName      string `json:"app.kubernetes.io/name,omitempty"`
//	AppKubernetesIoPartOf    string `json:"app.kubernetes.io/part-of,omitempty"`
//	AppKubernetesIoVersion   string `json:"app.kubernetes.io/version,omitempty"`
//	KymaProjectIoModule      string `json:"kyma-project.io/module,omitempty"`
//}

//type Metadata struct {
//	Annotations Annotations `json:"annotations,omitempty"`
//	Labels      Labels      `json:"labels,omitempty"`
//	Name        string      `json:"name,omitempty"`
//}
