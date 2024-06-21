package communitymodules

type Modules []Module

type Module struct {
	Name       string    `json:"name,omitempty"`
	Manageable bool      `json:"manageable,omitempty"`
	Versions   []Version `json:"versions,omitempty"`
}

type Version struct {
	Version      string     `json:"version,omitempty"`
	Channels     []string   `json:"channels,omitempty"`
	ManagerPath  string     `json:"managerPath,omitempty"`
	ManagerImage string     `json:"managerImage,omitempty"`
	Repository   string     `json:"repository,omitempty"`
	Resources    []Resource `json:"resources,omitempty"`
}

type Resource struct {
	Spec Spec `json:"spec,omitempty"`
}

type Spec struct {
	Group       string       `json:"group,omitempty"`
	Names       Names        `json:"names,omitempty"`
	Scope       string       `json:"scope,omitempty"`
	ApiVersions []ApiVersion `json:"versions,omitempty"`
}

type Names struct {
	Kind     string `json:"kind,omitempty"`
	ListKind string `json:"listKind,omitempty"`
	Plural   string `json:"plural,omitempty"`
	Singular string `json:"singular,omitempty"`
}

type ApiVersion struct {
	Name string `json:"name,omitempty"`
}
