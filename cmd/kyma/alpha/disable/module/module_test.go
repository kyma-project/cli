package module

import (
	"fmt"
	"github.com/kyma-project/lifecycle-manager/api/v1beta1"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDisableModule(t *testing.T) {
	installedModules := []v1beta1.Module{
		{
			"module1",
			"",
			"alpha",
			"Ignore",
		},
		{
			"module2",
			"",
			"",
			"CreateAndDelete",
		},
		{
			"module3",
			"",
			"regular",
			"",
		},
	}

	testCases := []struct {
		name            string
		existingModules []v1beta1.Module
		moduleName      string
		channel         string
		out             func(t *testing.T, outputModules []v1beta1.Module, err error)
	}{
		{
			name:            "Not found module",
			existingModules: installedModules,
			moduleName:      "module1",
			channel:         "regular",
			out: func(t *testing.T, outputModules []v1beta1.Module, err error) {
				require.Equal(t,
					installedModules,
					outputModules,
				)

				require.Equal(t, fmt.Errorf("could not disable module as it was not found: module1 in channel regular"), err)
			},
		},
		{
			name:            "Module disabled successfully from the module channel",
			existingModules: installedModules,
			moduleName:      "module3",
			channel:         "regular",
			out: func(t *testing.T, outputModules []v1beta1.Module, err error) {
				require.Equal(t,
					[]v1beta1.Module{
						{
							"module1",
							"",
							"alpha",
							"Ignore",
						},
						{
							"module2",
							"",
							"",
							"CreateAndDelete",
						},
					},
					outputModules,
				)

				require.Equal(t, nil, err)
			},
		},
		{
			name:            "Module disabled successfully from the global Kyma channel",
			existingModules: installedModules,
			moduleName:      "module2",
			channel:         "alpha",
			out: func(t *testing.T, outputModules []v1beta1.Module, err error) {
				require.Equal(t,
					[]v1beta1.Module{
						{
							"module1",
							"",
							"alpha",
							"Ignore",
						},
						{
							"module3",
							"",
							"regular",
							"",
						},
					},
					outputModules,
				)

				require.Equal(t, nil, err)
			},
		},
	}

	for _, test := range testCases {
		t.Run(
			test.name, func(t *testing.T) {
				modules, err := disableModule(test.existingModules, test.moduleName, test.channel)

				test.out(t, modules, err)
			},
		)
	}
}
