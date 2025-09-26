package types

import "strconv"

type NullableBool struct {
	Value *bool
}

func (n *NullableBool) String() string {
	if n.Value == nil {
		return ""
	}
	return strconv.FormatBool(*n.Value)
}

func (n *NullableBool) SetValue(value *string) error {
	if value == nil {
		return nil
	}

	b, err := strconv.ParseBool(*value)
	if err != nil {
		return err
	}
	n.Value = &b
	return nil
}

func (n *NullableBool) Set(value string) error {
	if value == "" {
		return nil
	}

	return n.SetValue(&value)
}

func (n *NullableBool) Type() string {
	return "bool"
}
