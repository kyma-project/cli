package types

type Scope string

const (
	ClusterScope   Scope = "cluster"
	NamespaceScope Scope = "namespace"
)

type ResourceInfo struct {
	Scope   Scope  `yaml:"scope"`
	Kind    string `yaml:"kind"`
	Group   string `yaml:"group"`
	Version string `yaml:"version"`
	// Singular string `yaml:"singular"`
	// Plural   string `yaml:"plural"`
}
