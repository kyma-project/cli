package modulesv2_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/entities"
	modulesfake "github.com/kyma-project/cli.v3/internal/modulesv2/fake"
	"github.com/stretchr/testify/require"
)

func TestPullService_Run(t *testing.T) {
	tests := []struct {
		name                        string
		pullConfig                  *dtos.PullConfig
		listExternalCommunityResult []*entities.ExternalModuleTemplate
		listExternalCommunityError  error
		getLocalCommunityResult     *entities.CommunityModuleTemplate
		getLocalCommunityError      error
		saveCommunityError          error

		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:                       "External community modules call fails",
			pullConfig:                 dtos.NewPullConfig("sample-module", "default", "", "", false),
			listExternalCommunityError: errors.New("moduleTemplatesRepository.ListExternalCommunity#Error"),
			expectedError:              true,
			expectedErrorMsg:           "failed to get community module from remote: failed to list external modules: moduleTemplatesRepository.ListExternalCommunity#Error",
		},
		{
			name:                        "Community module does not exist in external repo",
			pullConfig:                  dtos.NewPullConfig("sample-module", "default", "", "", false),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedError:               true,
			expectedErrorMsg:            "failed to get community module from remote: community module sample-module does not exist in the https://kyma-project.github.io/community-modules/all-modules.json repository",
		},
		{
			name:       "Operation fails to save moduletemplate in kyma-system namespace",
			pullConfig: dtos.NewPullConfig("sample-module", "kyma-system", "", "", false),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
			},
			expectedError:    true,
			expectedErrorMsg: "failed to store community module in the provided namespace: 'kyma-system' namespace is not allowed",
		},
		{
			name:       "Operation fails when provided version is not present in the repository",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "2.3.4", false),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			expectedError:    true,
			expectedErrorMsg: "failed to get community module from remote: community module sample-external-community-module:2.3.4 does not exist in remote repository",
		},
		{
			name:       "Operation fails when community module already exists and operation is not forced",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1", false),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			getLocalCommunityResult: modulesfake.CommunityModuleTemplate(&modulesfake.CommunityParams{Version: "1.0.1"}),
			expectedError:           true,
			expectedErrorMsg:        "failed to apply module template, 'sample-external-community-template-0.0.1' template already exists in the 'default' namespace. Use `--force` flag to override it",
		},
		{
			name:       "Operation fails on unsuccessful module save",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1", false),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			saveCommunityError: errors.New("moduleTemplatesRepository.SaveCommunityModule#Error"),
			expectedError:      true,
			expectedErrorMsg:   "moduleTemplatesRepository.SaveCommunityModule#Error",
		},
		{
			name:       "Operation succeeds",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1", false),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockModulesRepo := &modulesfake.ModuleTemplatesRepository{
				ListExternalCommunityResult: test.listExternalCommunityResult,
				ListExternalCommunityError:  test.listExternalCommunityError,
				GetLocalCommunityResult:     test.getLocalCommunityResult,
				GetLocalCommunityError:      test.getLocalCommunityError,
				SaveCommunityModuleError:    test.saveCommunityError,
			}

			service := modulesv2.NewPullService(mockModulesRepo)

			err := service.Run(context.Background(), test.pullConfig)

			if test.expectedError {
				require.Error(t, err)
				require.Equal(t, test.expectedErrorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
