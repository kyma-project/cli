package types

import (
	"fmt"
	"strings"
)

type MapElem struct {
	key string
	m   *Map
}

func NewMapElem(key string, m *Map) *MapElem {
	return &MapElem{
		key: key,
		m:   m,
	}
}

// Set implements the flag.Value interface
func (me *MapElem) Set(val string) error {
	if me.m.Values == nil {
		me.m.Values = map[string]interface{}{}
	}
	me.m.Values[me.key] = val
	return nil
}

// String implements the flag.Value interface
func (me *MapElem) String() string {
	if me.m == nil || me.m.Values == nil {
		return ""
	}
	if v, ok := me.m.Values[me.key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// Type implements the pflag.Value interface
func (me *MapElem) Type() string {
	return "string"
}

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

// SetValue sets the value of the map from a string in the format KEY=VALUE
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

	key := elems[0]
	if _, exists := em.Values[key]; exists {
		return fmt.Errorf("duplicate key: '%s' is provided multiple times", key)
	}

	em.Values[key] = elems[1]
	return nil
}

// Set implements the flag.Value interface
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
