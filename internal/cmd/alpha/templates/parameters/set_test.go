package parameters

import (
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSet(t *testing.T) {
	type args struct {
		obj    map[string]interface{}
		values []Value
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr clierror.Error
	}{
		{
			name: "set simple values",
			args: args{
				obj: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test",
					},
				},
				values: []Value{
					newStringValue(".metadata.name", "test-2"),       // replace
					newStringValue(".metadata.namespace", "default"), // set with existing .metadata
					newInt64Value(".spec.runtimes", 3),               // set new field
					newInt64Value(".spec.elems[].iter", 1),           // set slice
					newInt64Value(".spec.elems[].iter", 2),           // append slice
				},
			},
			want: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test-2",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"runtimes": int64(3),
					"elems": []interface{}{
						map[string]interface{}{
							"iter": int64(2),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "different fields kinds",
			args: args{
				obj: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test",
					},
				},
				values: []Value{
					newInt64Value(".metadata.name", 1), // should be string
				},
			},
			want: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
			wantErr: clierror.Wrap(errors.New("fields have different types for key name: type int64 other than expected string"),
				clierror.New("failed to set value 1 for path .metadata.name"),
			),
		},
		{
			name: "set slice",
			args: args{
				obj: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test",
					},
				},
				values: []Value{
					&fakeValue{
						value: []interface{}{"1", "2", "3"}, // int is not supported
						path:  ".spec.elems",
					},
				},
			},
			want: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test",
				},
				"spec": map[string]interface{}{
					"elems": []interface{}{
						"1",
						"2",
						"3",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "merge slices",
			args: args{
				obj: map[string]interface{}{
					"spec": map[string]interface{}{
						"elems": []interface{}{
							"1",
							"2",
							[]interface{}{},
						},
					},
				},
				values: []Value{
					&fakeValue{
						value: []interface{}{int64(1), "2", []interface{}{"1a"}, "12", "14"}, // int is not supported
						path:  ".spec.elems",
					},
				},
			},
			want: map[string]interface{}{
				"spec": map[string]interface{}{
					"elems": []interface{}{
						int64(1),
						"2",
						[]interface{}{"1a"},
						"12",
						"14",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &unstructured.Unstructured{Object: tt.args.obj}
			err := Set(u, tt.args.values)
			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.want, u.Object)
		})
	}
}

type fakeValue struct {
	value interface{}
	path  string
}

func (v *fakeValue) GetValue() interface{} {
	return v.value
}

func (v *fakeValue) GetPath() string {
	return v.path
}

func (v *fakeValue) Set(_ string) error {
	return nil
}

func (v *fakeValue) String() string {
	return ""
}

func (v *fakeValue) Type() string {
	return ""
}
