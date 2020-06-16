package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReplaceDockerImageURL(t *testing.T) {
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
		err := replaceInstallerImage(tt.data, replacedWithData)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, tt.data, tt.expectedResult, tt.testName)
		} else {
			require.NotNil(t, err, tt.testName)
		}
	}
}
