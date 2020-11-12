package installation

import (
	"encoding/base64"
	"path"
	"testing"

	stepMocks "github.com/kyma-project/cli/pkg/step/mocks"
	"github.com/stretchr/testify/require"
)

func Test_GetMasterHash(t *testing.T) {
	t.Parallel()
	h, err := getMasterHash()
	require.NoError(t, err)
	require.True(t, isHex(h))
}

func Test_GetLatestAvailableMasterHash(t *testing.T) {
	t.Parallel()
	h, err := getLatestAvailableMasterHash(&stepMocks.Step{}, 5)
	require.NoError(t, err)
	require.True(t, isHex(h))
}

func Test_LoadInstallationFiles(t *testing.T) {
	t.Parallel()
	localInstallation := Installation{
		Options: &Options{
			IsLocal:          true,
			fromLocalSources: false,
			configVersion:    "master-6dba1d2c",
			bucket:           developmentBucket,
		},
	}

	m, err := localInstallation.loadInstallationFiles()
	require.NoError(t, err)
	require.Equal(t, 3, len(m))

	clusterInstallation := Installation{
		Options: &Options{
			IsLocal:          false,
			fromLocalSources: false,
			configVersion:    "master-6dba1d2c",
			bucket:           developmentBucket,
		},
	}

	m, err = clusterInstallation.loadInstallationFiles()
	require.NoError(t, err)
	require.Equal(t, 2, len(m))

	f, err := loadStringContent(m)
	require.NoError(t, err)
	require.Equal(t, 2, len(f))
}

func Test_LoadConfigurations(t *testing.T) {
	t.Parallel()
	domain := "test.kyma"
	tlsCert := "testCert"
	tlsKey := "testKey"
	password := "testPass"

	installation := &Installation{
		Options: &Options{
			OverrideConfigs: []string{path.Join("../../internal/testdata", "overrides.yaml")},
			IsLocal:         false,
			Domain:          domain,
			TLSCert:         tlsCert,
			TLSKey:          tlsKey,
			Password:        password,
		},
	}

	configurations, err := installation.loadConfigurations(nil)
	require.NoError(t, err)
	require.Equal(t, 4, len(configurations.Configuration))
	dom, ok := configurations.Configuration.Get("global.domainName")
	require.Equal(t, true, ok)
	require.Equal(t, domain, dom.Value)
	tlsC, ok := configurations.Configuration.Get("global.tlsCrt")
	require.Equal(t, true, ok)
	require.Equal(t, tlsCert, tlsC.Value)
	tlsK, ok := configurations.Configuration.Get("global.tlsKey")
	require.Equal(t, true, ok)
	require.Equal(t, tlsKey, tlsK.Value)
	pass, ok := configurations.Configuration.Get("global.adminPassword")
	require.Equal(t, true, ok)
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte(password)), pass.Value)

	require.Equal(t, 1, len(configurations.ComponentConfiguration))
	require.Equal(t, "ory", configurations.ComponentConfiguration[0].Component)
	cpuR, ok := configurations.ComponentConfiguration[0].Configuration.Get("hydra.deployment.resources.requests.cpu")
	require.Equal(t, true, ok)
	require.Equal(t, "53m", cpuR.Value)
	cpuL, ok := configurations.ComponentConfiguration[0].Configuration.Get("hydra.deployment.resources.limits.cpu")
	require.Equal(t, true, ok)
	require.Equal(t, "153m", cpuL.Value)
}

func Test_LoadComponentsConfig(t *testing.T) {
	t.Parallel()
	installation := &Installation{
		Options: &Options{
			ComponentsConfig: path.Join("../../internal/testdata", "components.yaml"),
		},
	}

	components, err := LoadComponentsConfig(installation.Options.ComponentsConfig)
	require.NoError(t, err)
	require.Equal(t, 6, len(components))

	installation2 := &Installation{
		Options: &Options{
			ComponentsConfig: path.Join("../../internal/testdata", "installationCR.yaml"),
		},
	}

	components, err = LoadComponentsConfig(installation2.Options.ComponentsConfig)
	require.NoError(t, err)
	require.Equal(t, 8, len(components))
}

func Test_GetInstallerImage(t *testing.T) {
	t.Parallel()
	const image = "eu.gcr.io/kyma-project/kyma-installer:63f27f76"
	testData := File{Content: []map[string]interface{}{{
		"apiVersion": "installer.kyma-project.io/v1alpha1",
		"kind":       "Deployment",
		"spec": map[interface{}]interface{}{
			"template": map[interface{}]interface{}{
				"spec": map[interface{}]interface{}{
					"serviceAccountName": "kyma-installer",
					"containers": []interface{}{
						map[interface{}]interface{}{
							"name":  "kyma-installer-container",
							"image": image,
						},
					},
				},
			},
		},
	},
	},
	}

	insImage, err := getInstallerImage(&testData)
	require.NoError(t, err)
	require.Equal(t, image, insImage)
}

func Test_ReplaceDockerImageURL(t *testing.T) {
	t.Parallel()
	const replacedWithData = "testImage!"
	testData := []struct {
		testName       string
		data           File
		expectedResult File
		shouldFail     bool
	}{
		{
			testName: "correct data test",
			data: File{Content: []map[string]interface{}{{
				"apiVersion": "installer.kyma-project.io/v1alpha1",
				"kind":       "Deployment",
				"spec": map[interface{}]interface{}{
					"template": map[interface{}]interface{}{
						"spec": map[interface{}]interface{}{
							"serviceAccountName": "kyma-installer",
							"containers": []interface{}{
								map[interface{}]interface{}{
									"name":  "kyma-installer-container",
									"image": "eu.gcr.io/kyma-project/kyma-installer:63f27f76",
								},
							},
						},
					},
				},
			},
			},
			},
			expectedResult: File{Content: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Deployment",
					"spec": map[interface{}]interface{}{
						"template": map[interface{}]interface{}{
							"spec": map[interface{}]interface{}{
								"serviceAccountName": "kyma-installer",
								"containers": []interface{}{
									map[interface{}]interface{}{
										"name":  "kyma-installer-container",
										"image": replacedWithData,
									},
								},
							},
						},
					},
				},
			},
			},
			shouldFail: false,
		},
	}

	for _, tt := range testData {
		err := replaceInstallerImage(&tt.data, replacedWithData)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, tt.data, tt.expectedResult, tt.testName)
		} else {
			require.NotNil(t, err, tt.testName)
		}
	}
}

func Test_IsDockerImage(t *testing.T) {
	t.Parallel()
	ok := isDockerImage("testRegistry/testImage:tag")
	require.True(t, ok)

	ok = isDockerImage("testImage")
	require.False(t, ok)
}

func Test_IsSemVer(t *testing.T) {
	t.Parallel()
	ok := isSemVer("1.2.3")
	require.True(t, ok)

	ok = isSemVer("testVersion")
	require.False(t, ok)

	ok = isSemVer("12345")
	require.False(t, ok)
}
