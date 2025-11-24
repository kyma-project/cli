package entities

type ClusterMetadata struct {
	IsManagedByKLM bool
}

func NewClusterMetadata(isManagedByKLM bool) *ClusterMetadata {
	return &ClusterMetadata{isManagedByKLM}
}
