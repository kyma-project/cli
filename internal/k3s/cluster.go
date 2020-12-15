package k3s

import "encoding/json"

type clusterList struct {
	clusters []cluster
}

type cluster struct {
	name  string
	nodes []node
}

type node struct {
	name   string
	role   string
	labels map[string]string
	state  state
}

type state struct {
	running bool
	status  string
}

func (cl *clusterList) Unmarshal(data []byte) error {
	var clusters []cluster
	if err := json.Unmarshal(data, &clusters); err != nil {
		return err
	}
	cl.clusters = clusters
	return nil
}
