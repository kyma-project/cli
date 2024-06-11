package model

type Module []struct {
	Name       string     `json:"name,omitempty"`
	Manageable bool       `json:"manageable,omitempty"`
	Versions   []Versions `json:"versions,omitempty"`
}

type Versions struct {
	Version      string      `json:"version,omitempty"`
	Channels     []string    `json:"channels,omitempty"`
	ManagerPath  string      `json:"managerPath,omitempty"`
	ManagerImage string      `json:"managerImage,omitempty"`
	Resources    []Resources `json:"resources,omitempty"`
}

type Resources struct {
	Spec Spec `json:"spec,omitempty"`
}

type Spec struct {
	Group       string        `json:"group,omitempty"`
	Names       Names         `json:"names,omitempty"`
	Scope       string        `json:"scope,omitempty"`
	ApiVersions []ApiVersions `json:"versions,omitempty"`
}

type Names struct {
	Kind     string `json:"kind,omitempty"`
	ListKind string `json:"listKind,omitempty"`
	Plural   string `json:"plural,omitempty"`
	Singular string `json:"singular,omitempty"`
}

type ApiVersions struct {
	Name string `json:"name,omitempty"`
}
