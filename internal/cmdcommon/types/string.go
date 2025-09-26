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

func (n *NullableString) SetValue(value *string) error {
	n.Value = value
	return nil
}

func (n *NullableString) Set(value string) error {
	if value == "" {
		return nil
	}

	return n.SetValue(&value)
}

func (n *NullableString) Type() string {
	return "string"
}
