package prompt_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
)

func TestBoolPrompt_Yes(t *testing.T) {
	b := prompt.NewBool("Proceed?", false)
	withInput("y\n", func() {
		result, err := b.Prompt()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !result {
			t.Errorf("expected true for 'y' input, got false")
		}
	})
}

func TestBoolPrompt_No(t *testing.T) {
	b := prompt.NewBool("Proceed?", true)
	withInput("n\n", func() {
		result, err := b.Prompt()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result {
			t.Errorf("expected false for 'n' input, got true")
		}
	})
}

func TestBoolPrompt_DefaultValueTrue(t *testing.T) {
	b := prompt.NewBool("Proceed?", true)
	withInput("\n", func() {
		result, err := b.Prompt()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !result {
			t.Errorf("expected true for '' input and defaultValue = true, got false")
		}
	})
}

func TestBoolPrompt_DefaultValueFalse(t *testing.T) {
	b := prompt.NewBool("Proceed?", false)
	withInput("\n", func() {
		result, err := b.Prompt()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result {
			t.Errorf("expected false for '' input and defaultValue = false, got true")
		}
	})
}

func TestBoolPrompt_Invalid(t *testing.T) {
	b := prompt.NewBool("Proceed?", false)
	withInput("maybe\n", func() {
		_, err := b.Prompt()
		if err == nil {
			t.Error("expected error for invalid input, got nil")
		}
		if !strings.Contains(err.Error(), "invalid input") {
			t.Errorf("expected invalid input error, got: %v", err)
		}
	})
}

func withInput(input string, fn func()) {
	old := os.Stdin
	r, w, _ := os.Pipe()
	_, err := w.Write([]byte(input))
	w.Close()
	if err != nil {
		fmt.Println("writing error")
		return
	}

	os.Stdin = r
	defer func() { os.Stdin = old }()
	fn()
}
