package entities

type ClusterMetadata struct {
	IsManagedByKLM           bool
	ModuleTemplateCRDVersion string
}

func NewClusterMetadata(isManagedByKLM bool, moduleTemplateCRDVersion string) *ClusterMetadata {
	return &ClusterMetadata{isManagedByKLM, moduleTemplateCRDVersion}
}
