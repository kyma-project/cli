package install

import (
	"reflect"
	"testing"
)

func Test_RemoveActionLabel(t *testing.T) {
	testData := []struct {
		data           []map[string]interface{}
		expectedResult []map[string]interface{}
		shouldFail     bool
	}{
		{
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
				t.Fatal("Test failed but it is expected not to: ", err.Error())
			}
		}

	}
}
