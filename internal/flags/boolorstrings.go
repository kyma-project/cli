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
		// Enable the flag but do not clear existing values to preserve
		// earlier URLs when a later bare --remote is provided.
		r.Enabled = true
		return nil
	}

	if value == "false" {
		r.Enabled = false
		r.Values = nil
		return nil
	}

	r.Enabled = true
	// Support repeated flags and comma-separated lists; trim spaces and ignore empties
	parts := strings.SplitSeq(value, ",")
	for p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			r.Values = append(r.Values, v)
		}
	}
	return nil
}

func (r *BoolOrStrings) Type() string {
	return "bool or []string"
}
