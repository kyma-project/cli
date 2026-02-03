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
			pullConfig:                 dtos.NewPullConfig("sample-module", "default", "", ""),
			listExternalCommunityError: errors.New("moduleTemplatesRepository.ListExternalCommunity#Error"),
			expectedError:              true,
			expectedErrorMsg:           "failed to get community module from remote: failed to list external modules: moduleTemplatesRepository.ListExternalCommunity#Error",
		},
		{
			name:                        "Community module does not exist in external repo",
			pullConfig:                  dtos.NewPullConfig("sample-module", "default", "", ""),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{},
			expectedError:               true,
			expectedErrorMsg:            "failed to get community module from remote: community module sample-module does not exist in the https://kyma-project.github.io/community-modules/all-modules.json repository",
		},
		{
			name:       "Operation fails to save moduletemplate in kyma-system namespace",
			pullConfig: dtos.NewPullConfig("sample-module", "kyma-system", "", ""),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
			},
			expectedError:    true,
			expectedErrorMsg: "failed to store community module in the provided namespace: 'kyma-system' namespace is not allowed",
		},
		{
			name:       "Operation fails when provided version is not present in the repository",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "2.3.4"),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			expectedError:    true,
			expectedErrorMsg: "failed to get community module from remote: community module sample-external-community-module:2.3.4 does not exist in remote repository",
		},
		{
			name:       "Operation fails on unsuccessful module save",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1"),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.0"}),
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			saveCommunityError: errors.New("moduleTemplatesRepository.SaveCommunityModule#Error"),
			expectedError:      true,
			expectedErrorMsg:   "failed to save sample-external-community-template-0.0.1 module template in the default namespace: moduleTemplatesRepository.SaveCommunityModule#Error",
		},
		{
			name:       "Operation succeeds",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1"),
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

			_, err := service.Run(context.Background(), test.pullConfig)

			if test.expectedError {
				require.Error(t, err)
				require.Equal(t, test.expectedErrorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPullService_GetInstalledModuleTemplate(t *testing.T) {
	tests := []struct {
		name                        string
		pullConfig                  *dtos.PullConfig
		listExternalCommunityResult []*entities.ExternalModuleTemplate
		listExternalCommunityError  error
		getLocalCommunityResult     *entities.CommunityModuleTemplate
		getLocalCommunityError      error

		expectedResult   *dtos.PullResult
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:                       "External community modules call fails",
			pullConfig:                 dtos.NewPullConfig("sample-module", "default", "", ""),
			listExternalCommunityError: errors.New("moduleTemplatesRepository.ListExternalCommunity#Error"),
			expectedError:              true,
			expectedErrorMsg:           "failed to get community module from remote: failed to list external modules: moduleTemplatesRepository.ListExternalCommunity#Error",
		},
		{
			name:       "GetLocalCommunity call fails",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1"),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			getLocalCommunityError: errors.New("moduleTemplatesRepository.GetLocalCommunity#Error"),
			expectedError:          true,
			expectedErrorMsg:       "failed to get community module from the target kyma cluster: moduleTemplatesRepository.GetLocalCommunity#Error",
		},
		{
			name:       "Module exists locally - returns PullResult",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1"),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			getLocalCommunityResult: modulesfake.CommunityModuleTemplate(&modulesfake.CommunityParams{
				TemplateName: "sample-community-template-1.0.1",
				ModuleName:   "sample-community-module",
				Version:      "1.0.1",
				Namespace:    "default",
			}),
			expectedResult: &dtos.PullResult{
				ModuleName:         "sample-community-module",
				ModuleTemplateName: "sample-community-template-1.0.1",
				Version:            "1.0.1",
				Namespace:          "default",
			},
			expectedError: false,
		},
		{
			name:       "Module does not exist locally - returns nil",
			pullConfig: dtos.NewPullConfig("sample-module", "default", "", "1.0.1"),
			listExternalCommunityResult: []*entities.ExternalModuleTemplate{
				modulesfake.ExternalModuleTemplate(&modulesfake.ExternalParams{Version: "1.0.1"}),
			},
			getLocalCommunityResult: nil,
			expectedResult:          nil,
			expectedError:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockModulesRepo := &modulesfake.ModuleTemplatesRepository{
				ListExternalCommunityResult: test.listExternalCommunityResult,
				ListExternalCommunityError:  test.listExternalCommunityError,
				GetLocalCommunityResult:     test.getLocalCommunityResult,
				GetLocalCommunityError:      test.getLocalCommunityError,
			}

			service := modulesv2.NewPullService(mockModulesRepo)

			result, err := service.GetInstalledModuleTemplate(context.Background(), test.pullConfig)

			if test.expectedError {
				require.Error(t, err)
				require.Equal(t, test.expectedErrorMsg, err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedResult, result)
			}
		})
	}
}
