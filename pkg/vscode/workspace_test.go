package vscode

import (
	"testing"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
)

func Test_it(t *testing.T) {
	if err := Workspace.build("/tmp/", workspace.DefaultWriterProvider); err != nil {
		t.Log(err)
	}
}
