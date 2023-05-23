package module

import (
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"reflect"
	"testing"
)

func Test_disableModule(t *testing.T) {
	installedModules := []v1beta2.Module{
		{
			Name:                 "module1",
			ControllerName:       "",
			Channel:              "alpha",
			CustomResourcePolicy: "Ignore",
		},
		{
			Name:                 "module2",
			ControllerName:       "",
			Channel:              "",
			CustomResourcePolicy: "CreateAndDelete",
		},
		{
			Name:                 "module3",
			ControllerName:       "",
			Channel:              "regular",
			CustomResourcePolicy: "",
		},
	}

	type args struct {
		modules []v1beta2.Module
		name    string
		channel string
	}
	tests := []struct {
		name    string
		args    args
		want    []v1beta2.Module
		wantErr bool
	}{
		{
			name: "Not found module",
			args: args{
				modules: installedModules,
				name:    "module1",
				channel: "regular",
			},
			want:    installedModules,
			wantErr: true,
		},
		{
			name: "Module disabled successfully from the module channel",
			args: args{
				modules: installedModules,
				name:    "module3",
				channel: "regular",
			},
			want: []v1beta2.Module{
				{
					Name:                 "module1",
					ControllerName:       "",
					Channel:              "alpha",
					CustomResourcePolicy: "Ignore",
				},
				{
					Name:                 "module2",
					ControllerName:       "",
					Channel:              "",
					CustomResourcePolicy: "CreateAndDelete",
				},
			},
			wantErr: false,
		},
		{
			name: "Module disabled successfully from the global Kyma channel",
			args: args{
				modules: installedModules,
				name:    "module2",
				channel: "alpha",
			},
			want: []v1beta2.Module{
				{
					Name:                 "module1",
					ControllerName:       "",
					Channel:              "alpha",
					CustomResourcePolicy: "Ignore",
				},
				{
					Name:                 "module3",
					ControllerName:       "",
					Channel:              "regular",
					CustomResourcePolicy: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := disableModule(tt.args.modules, tt.args.name, tt.args.channel)
			if (err != nil) != tt.wantErr {
				t.Errorf("disableModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("disableModule() got = %v, want %v", got, tt.want)
			}
		})
	}
}
