package dtos

type ListResult struct {
	Name                 string
	Version              string
	Channel              string
	ModuleState          string
	Managed              bool
	CustomResourcePolicy string
	InstallationState    string
}
