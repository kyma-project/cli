package common

import (
	"testing"
	"text/template"

	"github.com/kyma-project/cli.v3/internal/extensions/types"
)

func TestNewIfNil_ValueUsedWhenFlagNotNil(t *testing.T) {

	var overwrites types.ActionConfigOverwrites
	funcMap := template.FuncMap{
		"newLineIndent": newLineIndent,
		"toEnvs":        toEnvs,
		"toArray":       toArray,
		"toYaml":        toYaml,
	}
	ifNilRaw := newIfNil(overwrites, &funcMap)
	ifNil, ok := ifNilRaw.(func(interface{}, interface{}, interface{}) (interface{}, error))
	if !ok {
		t.Fatalf("unexpected ifNil type")
	}

	out, err := ifNil("hello", "fallback", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Fatalf("expected 'hello', got %v", out)
	}
}

func TestNewIfNil_ValueForNilUsedWhenFlagNilAndTemplateExecuted(t *testing.T) {
	var overwrites types.ActionConfigOverwrites
	funcMap := template.FuncMap{
		"newLineIndent": newLineIndent,
		"toEnvs":        toEnvs,
		"toArray":       toArray,
		"toYaml":        toYaml,
	}
	ifNilRaw := newIfNil(overwrites, &funcMap)
	ifNil := ifNilRaw.(func(interface{}, interface{}, interface{}) (interface{}, error))

	valForNil := `prefix${{ newLineIndent 2 "a\nb" }}suffix`
	out, err := ifNil("unused", valForNil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "prefixa\n  bsuffix"
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}

func TestNewIfNil_ParseError(t *testing.T) {
	var overwrites types.ActionConfigOverwrites
	funcMap := template.FuncMap{
		"newLineIndent": newLineIndent,
		"toEnvs":        toEnvs,
		"toArray":       toArray,
		"toYaml":        toYaml,
	}
	ifNil := newIfNil(overwrites, &funcMap).(func(interface{}, interface{}, interface{}) (interface{}, error))

	_, err := ifNil(`bad ${{ if }}`, "fallback", true)
	if err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}

func TestNewIfNil_NonStringValue(t *testing.T) {
	var overwrites types.ActionConfigOverwrites
	funcMap := template.FuncMap{}
	ifNil := newIfNil(overwrites, &funcMap).(func(interface{}, interface{}, interface{}) (interface{}, error))

	out, err := ifNil(42, 100, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != 42 {
		t.Fatalf("expected 42, got %v", out)
	}
}
