package install

import (
	"reflect"
	"testing"
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
	}

	cmd := &command{
		opts: nil,
	}

	for _, tt := range testData {
		err := cmd.removeActionLabel(tt.data)

		if !reflect.DeepEqual(tt.data, tt.expectedResult) {
			t.Fatalf("\r\nResult:\t%v\r\nExpected:\t%v\r\n", tt.data,
				tt.expectedResult)
		}

		if err != nil {
			if !tt.shouldFail {
				t.Fatal("Test expected to fail but it hasn't")
			}
		} else {
			if tt.shouldFail {
				t.Logf("\r\nResult:\t%v\r\nExpected:\t%v\r\n", tt.data,
					tt.expectedResult)
				t.Fatal("Test failed but it is expected not to")
			}
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
		if err != nil {
			if !tt.shouldFail {
				t.Fatalf("Test '%s' failed but it shouldn't. Error: %s\r\n",
					tt.testName, err.Error())
			}
		}
		if !reflect.DeepEqual(res, tt.expectedResult) {
			if !tt.shouldFail {
				t.Fatalf("\r\nExpected:\t%v\r\nReceived:\t%v\r\n", tt.expectedResult, res)
			}
		}
	}
}
