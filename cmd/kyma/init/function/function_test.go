package function

import (
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestFunctionFlags ensures that the provided command flags are stored in the options.
func TestFunctionFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, "", o.Name, "Default value for the --name flag not as expected.")
	require.Equal(t, "", o.Namespace, "Default value for the --namespace flag not as expected.")
	require.Equal(t, "", o.Dir, "Default value for the --dir flag not as expected.")
	require.Equal(t, "nodejs20", o.Runtime, "Default value for the --runtime flag not as expected.")
	require.Equal(t, "", o.RuntimeImageOverride, "The parsed value for the --runtime-image-override flag not as expected.")
	require.Equal(t, "", o.URL, "The parsed value for the --url flag not as expected.")
	require.Equal(t, "", o.RepositoryName, "The parsed value for the --repository-name flag not as expected.")
	require.Equal(t, "main", o.Reference, "The parsed value for the --reference flag not as expected.")
	require.Equal(t, "/", o.BaseDir, "The parsed value for the --base-dir flag not as expected.")
	require.Equal(t, false, o.VsCode, "Default value for the --vscode flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"--dir", "/fakepath",
		"--name", "test-name",
		"--namespace", "test-namespace",
		"--runtime-image-override", "runtime-image-override",
		"--runtime", "python312",
		"--url", "test-url",
		"--repository-name", "test-repository-name",
		"--reference", "test-reference",
		"--base-dir", "test-base-dir",
		"--vscode",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/fakepath", o.Dir, "The parsed value for the --dir flag not as expected.")
	require.Equal(t, "test-name", o.Name, "The parsed value for the --name flag not as expected.")
	require.Equal(t, "test-namespace", o.Namespace, "The parsed value for the --namespace flag not as expected.")
	require.Equal(t, "python312", o.Runtime, "The parsed value for the --runtime flag not as expected.")
	require.Equal(t, "runtime-image-override", o.RuntimeImageOverride, "The parsed value for the --runtime-image-override flag not as expected.")
	require.Equal(t, "test-url", o.URL, "The parsed value for the --url flag not as expected.")
	require.Equal(t, "test-repository-name", o.RepositoryName, "The parsed value for the --repository-name flag not as expected.")
	require.Equal(t, "test-reference", o.Reference, "The parsed value for the --reference flag not as expected.")
	require.Equal(t, "test-base-dir", o.BaseDir, "The parsed value for the --base-dir flag not as expected.")
	require.Equal(t, true, o.VsCode, "Parsed value for the --vscode flag not as expected.")

	err = c.ParseFlags([]string{
		"-d", "/tmpfile",
		"-r", "nodejs20",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/tmpfile", o.Dir, "The parsed value for the --dir flag not as expected.")
	require.Equal(t, "test-name", o.Name, "The parsed value for the --name flag not as expected.")
	require.Equal(t, "test-namespace", o.Namespace, "The parsed value for the --namespace flag not as expected.")
	require.Equal(t, "nodejs20", o.Runtime, "The parsed value for the --runtime flag not as expected.")
	require.Equal(t, "runtime-image-override", o.RuntimeImageOverride, "The parsed value for the --runtime-image-override flag not as expected.")
	require.Equal(t, "test-url", o.URL, "The parsed value for the --url flag not as expected.")
	require.Equal(t, "test-repository-name", o.RepositoryName, "The parsed value for the --repository-name flag not as expected.")
	require.Equal(t, "test-reference", o.Reference, "The parsed value for the --reference flag not as expected.")
	require.Equal(t, "test-base-dir", o.BaseDir, "The parsed value for the --base-dir flag not as expected.")
}
