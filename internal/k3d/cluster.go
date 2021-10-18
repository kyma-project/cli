package k3d

import "encoding/json"

//ClusterList containing cluster entities
type ClusterList struct {
	Clusters []Cluster
}

//Cluster including list of nodes
type Cluster struct {
	Name  string
	Nodes []Node
}

//Node in the K3s setup (could be lb, main, agent etc.)
type Node struct {
	Name  string
	State State
}

//State of a node
type State struct {
	Running bool
	Status  string
}

//Unmarshal converts a JSON to nested structs
func (cl *ClusterList) Unmarshal(data []byte) error {
	var clusters []Cluster
	if err := json.Unmarshal(data, &clusters); err != nil {
		return err
	}
	cl.Clusters = clusters
	return nil
}
