package prompt_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
)

func TestListPrompt_ValidInput(t *testing.T) {
	l := prompt.NewList("Select a fruit:", []string{"apple", "banana", "cherry"})
	withInput("banana\n", func() {
		result, err := l.Prompt()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result != "banana" {
			t.Errorf("expected 'banana', got '%s'", result)
		}
	})
}

func TestListPrompt_InvalidInput(t *testing.T) {
	l := prompt.NewList("Select a fruit:", []string{"apple", "banana", "cherry"})
	withInput("orange\n", func() {
		_, err := l.Prompt()
		if err == nil {
			t.Error("expected error for invalid input, got nil")
		}
	})
}
