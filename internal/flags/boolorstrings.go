package flags

import "strings"

type BoolOrStrings struct {
	Values  []string
	Enabled bool
}

func (r *BoolOrStrings) String() string {
	if len(r.Values) > 0 {
		return strings.Join(r.Values, ",")
	}

	if r.Enabled {
		return "true"
	}

	return "false"
}

func (r *BoolOrStrings) Set(value string) error {
	if value == "" || value == "true" {
		r.Enabled = true
		r.Values = nil
		return nil
	}

	if value == "false" {
		r.Enabled = false
		r.Values = nil
		return nil
	}

	r.Enabled = true
	r.Values = strings.Split(value, ",")
	return nil
}

func (r *BoolOrStrings) Type() string {
	return "bool or []string"
}

func (r *BoolOrStrings) IsBoolFlag() bool {
	return true
}
