package dtos

type ListResult struct {
	Name                 string
	Version              string
	Channel              string
	State                string
	Managed              bool
	CustomResourcePolicy string
	InstallationState    string
}
