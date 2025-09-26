package common

import (
	"strings"
	"testing"
	"text/template"

	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/stretchr/testify/assert"
)

func TestNewIfNil(t *testing.T) {
	var overwrites types.ActionConfigOverwrites
	funcMap := template.FuncMap{
		"newLineIndent": newLineIndent,
		"toEnvs":        toEnvs,
		"toArray":       toArray,
		"toYaml":        toYaml,
	}

	t.Run("value used when flag not nil", func(t *testing.T) {
		ifNilRaw := newIfNil(overwrites, &funcMap)
		ifNil, ok := ifNilRaw.(func(any, any, any) (any, error))
		assert.True(t, ok, "unexpected ifNil type")

		out, err := ifNil("hello", "fallback", true)
		assert.NoError(t, err)
		assert.Equal(t, "hello", out)
	})

	t.Run("value for nil used when flag nil and template executed", func(t *testing.T) {
		ifNilRaw := newIfNil(overwrites, &funcMap)
		ifNil := ifNilRaw.(func(any, any, any) (any, error))

		valForNil := `prefix${{ newLineIndent 2 "a\nb" }}suffix`
		out, err := ifNil("unused", valForNil, nil)
		assert.NoError(t, err)

		expected := "prefixa\n  bsuffix"
		assert.Equal(t, expected, out)
	})

	t.Run("parse error", func(t *testing.T) {
		ifNil := newIfNil(overwrites, &funcMap).(func(any, any, any) (any, error))

		_, err := ifNil(`bad ${{ if }}`, "fallback", true)
		assert.Error(t, err, "expected parse error")
	})

	t.Run("non string value", func(t *testing.T) {
		funcMapEmpty := template.FuncMap{}
		ifNil := newIfNil(overwrites, &funcMapEmpty).(func(any, any, any) (any, error))

		out, err := ifNil(42, 100, true)
		assert.NoError(t, err)
		assert.Equal(t, 42, out)
	})

	t.Run("template execution error", func(t *testing.T) {
		// Using a funcMap without required functions to cause execution error
		limitedFuncMap := template.FuncMap{}
		ifNil := newIfNil(overwrites, &limitedFuncMap).(func(any, any, any) (any, error))

		valForNil := `${{ nonExistentFunc "test" }}`
		_, err := ifNil("unused", valForNil, nil)
		assert.Error(t, err)
	})

	t.Run("complex template with multiple functions", func(t *testing.T) {
		ifNil := newIfNil(overwrites, &funcMap).(func(any, any, any) (any, error))

		complexTemplate := `${{ newLineIndent 4 "line1\nline2" }}`
		out, err := ifNil("unused", complexTemplate, nil)
		assert.NoError(t, err)
		assert.Contains(t, out.(string), "line1")
		assert.Contains(t, out.(string), "    line2")
	})

	t.Run("map as flag value", func(t *testing.T) {
		ifNil := newIfNil(overwrites, &funcMap).(func(any, any, any) (any, error))

		out, err := ifNil("primary", "fallback", map[string]string{"key": "value"})
		assert.NoError(t, err)
		assert.Equal(t, "primary", out)
	})
}

func TestTemplateConfig(t *testing.T) {
	t.Run("simple template execution", func(t *testing.T) {
		tmpl := []byte("Hello World")
		overwrites := types.ActionConfigOverwrites{}

		result, err := templateConfig(tmpl, overwrites)
		assert.Nil(t, err)
		assert.Equal(t, []byte("Hello World"), result)
	})

	t.Run("template with function", func(t *testing.T) {
		tmpl := []byte(`${{ newLineIndent 2 "line1\nline2" }}`)
		overwrites := types.ActionConfigOverwrites{}

		result, err := templateConfig(tmpl, overwrites)
		assert.Nil(t, err)
		assert.Equal(t, "line1\n  line2", string(result))
	})

	t.Run("template parse error", func(t *testing.T) {
		tmpl := []byte(`${{ invalid template syntax`)
		overwrites := types.ActionConfigOverwrites{}

		result, err := templateConfig(tmpl, overwrites)
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.String(), "failed to parse config template")
	})

	t.Run("template execution error", func(t *testing.T) {
		tmpl := []byte(`${{ nonExistentFunction }}`)
		overwrites := types.ActionConfigOverwrites{}

		result, err := templateConfig(tmpl, overwrites)
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})
}

func TestNewLineIndent(t *testing.T) {
	t.Run("single line no change", func(t *testing.T) {
		result := newLineIndent(2, "hello")
		assert.Equal(t, "hello", result)
	})

	t.Run("multiple lines with indentation", func(t *testing.T) {
		result := newLineIndent(4, "line1\nline2\nline3")
		expected := "line1\n    line2\n    line3"
		assert.Equal(t, expected, result)
	})

	t.Run("zero indentation", func(t *testing.T) {
		result := newLineIndent(0, "line1\nline2")
		assert.Equal(t, "line1\nline2", result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := newLineIndent(2, "")
		assert.Equal(t, "", result)
	})

	t.Run("string with only newlines", func(t *testing.T) {
		result := newLineIndent(2, "\n\n")
		assert.Equal(t, "\n  \n  ", result)
	})
}

func TestToEnvs(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		envMap := map[string]interface{}{}
		result := toEnvs(envMap)
		assert.Equal(t, "[]", result)
	})

	t.Run("single environment variable", func(t *testing.T) {
		envMap := map[string]interface{}{
			"KEY1": "value1",
		}
		result := toEnvs(envMap)
		assert.Equal(t, `[{"name":"KEY1","value":"value1"}]`, result)
	})

	t.Run("multiple environment variables", func(t *testing.T) {
		envMap := map[string]interface{}{
			"KEY1": "value1",
			"KEY2": "value2",
		}
		result := toEnvs(envMap)

		// Since map iteration order is not guaranteed, check both possible orders
		expected1 := `[{"name":"KEY1","value":"value1"},{"name":"KEY2","value":"value2"}]`
		expected2 := `[{"name":"KEY2","value":"value2"},{"name":"KEY1","value":"value1"}]`

		assert.True(t, result == expected1 || result == expected2)
		assert.Contains(t, result, `"name":"KEY1"`)
		assert.Contains(t, result, `"value":"value1"`)
		assert.Contains(t, result, `"name":"KEY2"`)
		assert.Contains(t, result, `"value":"value2"`)
	})

	t.Run("environment variable with special characters", func(t *testing.T) {
		envMap := map[string]interface{}{
			"SPECIAL": "value with spaces and symbols!@#",
		}
		result := toEnvs(envMap)
		assert.Equal(t, `[{"name":"SPECIAL","value":"value with spaces and symbols!@#"}]`, result)
	})
}

func TestToArray(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		format := `"{{.key}}": "{{.value}}"`
		data := map[string]interface{}{}

		result, err := toArray(format, data)
		assert.NoError(t, err)
		assert.Equal(t, "[]", result)
	})

	t.Run("single item with simple format", func(t *testing.T) {
		format := `"{{.key}}": "{{.value}}"`
		data := map[string]interface{}{
			"name": "test",
		}

		result, err := toArray(format, data)
		assert.NoError(t, err)
		assert.Equal(t, `["name": "test"]`, result)
	})

	t.Run("multiple items", func(t *testing.T) {
		format := `"{{.key}}": "{{.value}}"`
		data := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}

		result, err := toArray(format, data)
		assert.NoError(t, err)

		// Check that result contains both items (order not guaranteed)
		assert.Contains(t, result, `"key1": "value1"`)
		assert.Contains(t, result, `"key2": "value2"`)
		assert.True(t, strings.HasPrefix(result, "["))
		assert.True(t, strings.HasSuffix(result, "]"))
	})

	t.Run("complex format template", func(t *testing.T) {
		format := `{"name": "{{.key}}", "value": "{{.value}}", "type": "string"}`
		data := map[string]interface{}{
			"config": "production",
		}

		result, err := toArray(format, data)
		assert.NoError(t, err)
		expected := `[{"name": "config", "value": "production", "type": "string"}]`
		assert.Equal(t, expected, result)
	})

	t.Run("invalid template format", func(t *testing.T) {
		format := `{{.invalid template`
		data := map[string]interface{}{
			"key": "value",
		}

		result, err := toArray(format, data)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to parse array template")
	})

	t.Run("template execution error", func(t *testing.T) {
		format := `{{.nonExistentField}}`
		data := map[string]interface{}{
			"key": "value",
		}

		result, err := toArray(format, data)
		assert.NoError(t, err) // Template execution with missing fields returns "<no value>" by default
		assert.Equal(t, "[<no value>]", result)
	})
}

func TestToYaml(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		data := map[string]interface{}{}
		result := toYaml(data)
		assert.Equal(t, "{}", result)
	})

	t.Run("single key-value pair", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
		}
		result := toYaml(data)
		assert.Equal(t, `{"name":"test"}`, result)
	})

	t.Run("multiple key-value pairs", func(t *testing.T) {
		data := map[string]interface{}{
			"name":    "test",
			"version": "1.0.0",
		}
		result := toYaml(data)

		// Check that result contains both pairs (order not guaranteed)
		assert.Contains(t, result, `"name":"test"`)
		assert.Contains(t, result, `"version":"1.0.0"`)
		assert.True(t, strings.HasPrefix(result, "{"))
		assert.True(t, strings.HasSuffix(result, "}"))
	})

	t.Run("values with special characters", func(t *testing.T) {
		data := map[string]interface{}{
			"description": "A test with spaces, numbers 123, and symbols!@#",
			"path":        "/home/user/file.txt",
		}
		result := toYaml(data)

		assert.Contains(t, result, `"description":"A test with spaces, numbers 123, and symbols!@#"`)
		assert.Contains(t, result, `"path":"/home/user/file.txt"`)
	})

	t.Run("numeric and boolean values", func(t *testing.T) {
		data := map[string]interface{}{
			"count":   42,
			"enabled": true,
			"rate":    3.14,
		}
		result := toYaml(data)

		// The toYaml function uses %s format, which produces Go's default string representation for non-strings
		assert.Contains(t, result, `"count":"%!s(int=42)"`)
		assert.Contains(t, result, `"enabled":"%!s(bool=true)"`)
		assert.Contains(t, result, `"rate":"%!s(float64=3.14)"`)
	})
}
