package types

import "strconv"

type NullableInt64 struct {
	Value *int64
}

func (n *NullableInt64) String() string {
	if n.Value == nil {
		return ""
	}
	return strconv.FormatInt(*n.Value, 10)
}

func (n *NullableInt64) SetValue(value *string) error {
	if value == nil {
		return nil
	}

	ni, err := strconv.ParseInt(*value, 10, 64)
	if err != nil {
		return err
	}
	n.Value = &ni
	return nil
}

func (n *NullableInt64) Set(value string) error {
	if value == "" {
		return nil
	}

	return n.SetValue(&value)
}

func (n *NullableInt64) Type() string {
	return "int"
}
