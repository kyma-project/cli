package cli

import "github.com/kyma-project/cli/internal/nice"

func AlphaWarn() {
	np := nice.Nice{}
	np.PrintImportant("WARNING: This command is experimental and might change in its final version. Use at your own risk.")
}
