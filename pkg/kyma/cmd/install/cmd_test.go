package install

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RemoveActionLabel(t *testing.T) {
	testData := []struct {
		testName       string
		data           []map[string]interface{}
		expectedResult []map[string]interface{}
		shouldFail     bool
	}{
		{
			testName: "correct data test",
			data: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name": "kyma-installation",
						"labels": map[interface{}]interface{}{
							"action": "install",
						},
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name":   "kyma-installation",
						"labels": map[interface{}]interface{}{},
					},
				},
			},
			shouldFail: false,
		},
		{
			testName: "incorrect data test",
			data: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name":   "kyma-installation",
						"labels": map[interface{}]interface{}{},
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					"apiVersion": "installer.kyma-project.io/v1alpha1",
					"kind":       "Installation",
					"metadata": map[interface{}]interface{}{
						"name":   "kyma-installation",
						"labels": map[interface{}]interface{}{},
					},
				},
			},
			shouldFail: true,
		},
	}

	cmd := &command{
		opts: nil,
	}

	for _, tt := range testData {
		err := cmd.removeActionLabel(tt.data)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, tt.data, tt.expectedResult, tt.testName)
		} else {
			require.Equal(t, tt.data, tt.expectedResult, tt.testName)
		}
	}
}

func Test_ReplaceDockerImageURL(t *testing.T) {
	const replacedWithData = "testImage!"
	testData := []struct {
		testName       string
		data           []map[string]interface{}
		expectedResult []map[string]interface{}
		shouldFail     bool
	}{
		{
			testName: "correct data test",
			data: []map[string]interface{}{
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
										"image": "eu.gcr.io/kyma-project/develop/kyma-installer:63f27f76",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: []map[string]interface{}{
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
			shouldFail: false,
		},
	}

	cmd := &command{
		opts: nil,
	}

	for _, tt := range testData {
		res, err := cmd.replaceDockerImageURL(tt.data, replacedWithData)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, res, tt.expectedResult, tt.testName)
		} else {
			require.NotNil(t, err, tt.testName)
		}
	}
}
