package k3d

import "encoding/json"

//RegistryList containing Registry entities
type RegistryList struct {
	Registries []Registry
}

//Registry including list of nodes
type Registry struct {
	Name  string
	State State
}

//Unmarshal converts a JSON to nested structs
func (cl *RegistryList) Unmarshal(data []byte) error {
	var registries []Registry
	if err := json.Unmarshal(data, &registries); err != nil {
		return err
	}
	cl.Registries = registries
	return nil
}
