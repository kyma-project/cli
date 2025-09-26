package types

import (
	"fmt"
	"strings"
)

type Map struct {
	Values map[string]interface{}
}

func (em *Map) String() string {
	values := make([]string, len(em.Values))

	index := 0
	for key, value := range em.Values {
		values[index] = fmt.Sprintf("%s=%s", key, value)
		index++
	}

	return strings.Join(values, ",")
}

func (em *Map) SetValue(value *string) error {
	if value == nil {
		return nil
	}

	if em.Values == nil {
		em.Values = map[string]interface{}{}
	}

	if *value == "" {
		// input is empty, do nothing
		return nil
	}

	elems := strings.Split(*value, "=")
	if len(elems) != 2 {
		return fmt.Errorf("failed to parse value '%s', should be in format KEY=VALUE", *value)
	}

	em.Values[elems[0]] = elems[1]
	return nil
}

func (em *Map) Set(value string) error {
	if value == "" {
		return nil
	}

	return em.SetValue(&value)
}

func (em *Map) Type() string {
	return "stringArray"
}

func (em *Map) GetNullableMap() map[string]*string {
	nullableMap := map[string]*string{}
	for key, value := range em.Values {
		v, _ := value.(string)
		nullableMap[key] = &v
	}

	return nullableMap
}
