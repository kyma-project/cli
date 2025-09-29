package parameters

import (
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
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
					// replace
					&stringValue{path: ".metadata.name",
						NullableString: types.NullableString{Value: ptr.To("test-2")}},
					// set with existing .metadata
					&pathValue{stringValue: stringValue{path: ".metadata.namespace",
						NullableString: types.NullableString{Value: ptr.To("default")}}},
					// set new field
					&int64Value{path: ".spec.runtimes",
						NullableInt64: types.NullableInt64{Value: ptr.To[int64](3)}},
					// set slice
					&int64Value{path: ".spec.elems[].iter",
						NullableInt64: types.NullableInt64{Value: ptr.To[int64](1)}},
					// overwrite slice
					&int64Value{path: ".spec.elems[].iter",
						NullableInt64: types.NullableInt64{Value: ptr.To[int64](2)}},
					// append another slice elem
					&boolValue{path: ".spec.elems[1].required",
						NullableBool: types.NullableBool{Value: ptr.To(true)}},
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
						map[string]interface{}{
							"required": true,
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
					&int64Value{path: ".metadata.name",
						NullableInt64: types.NullableInt64{Value: ptr.To[int64](1)}}, // should be string
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
			err := Set(u.Object, tt.args.values)
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

func (v *fakeValue) SetValue(_ *string) error {
	return nil
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
