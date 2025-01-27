package types

type NullableString struct {
	Value *string
}

func (n *NullableString) String() string {
	if n.Value == nil {
		return ""
	}
	return *n.Value
}

func (n *NullableString) Set(value string) error {
	if value == "" {
		return nil
	}

	n.Value = &value
	return nil
}

func (n *NullableString) Type() string {
	return "string"
}
