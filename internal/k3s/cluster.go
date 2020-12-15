package k3s

import "encoding/json"

//ClusterList containting cluster entites
type ClusterList struct {
	Clusters []Cluster
}

//Cluster including list of nodes
type Cluster struct {
	Name  string
	Nodes []Node
}

//Node in the K3s setup (could be lb, master, agent etc.)
type Node struct {
	Name   string
	Role   string
	Labels map[string]string
	State  State
}

//State of a node
type State struct {
	Running bool
	Status  string
}

//Unmarshal converst a JSON to nested structs
func (cl *ClusterList) Unmarshal(data []byte) error {
	var clusters []Cluster
	if err := json.Unmarshal(data, &clusters); err != nil {
		return err
	}
	cl.Clusters = clusters
	return nil
}
