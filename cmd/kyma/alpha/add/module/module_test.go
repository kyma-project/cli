package module

import (
	"context"
	"testing"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cli/cmd/kyma/alpha/add/module/mock"
	"github.com/kyma-project/cli/internal/cli/alpha/module"
)

func Test_validateChannel(t *testing.T) {

	ctx := context.TODO()
	filteredTemplates := []v1beta2.ModuleTemplate{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "test-module-A"},
			Spec:       v1beta2.ModuleTemplateSpec{Channel: "regular"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "test-module-A"},
			Spec:       v1beta2.ModuleTemplateSpec{Channel: "fast"},
		},
	}

	moduleInteractor := mock.Interactor{}
	moduleInteractor.Test(t)
	moduleInteractor.On("GetFilteredModuleTemplates", ctx).Return(filteredTemplates, nil)
	type args struct {
		ctx              context.Context
		moduleInteractor module.Interactor
		moduleIdentifier string
		channel          string
		kymaChannel      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Invalid Channel throws an error.",
			args: args{
				ctx:              ctx,
				moduleInteractor: &moduleInteractor,
				moduleIdentifier: "sample-module",
				channel:          "invalid",
				kymaChannel:      "regular",
			},
			wantErr: true,
		},
		{
			name: "Valid Module validation with specified channel.",
			args: args{
				ctx:              ctx,
				moduleInteractor: &moduleInteractor,
				moduleIdentifier: "sample-module",
				channel:          "fast",
				kymaChannel:      "regular",
			},
			wantErr: false,
		},
		{
			name: "Valid Module validation without specified channel, use Kyma channel",
			args: args{
				ctx:              ctx,
				moduleInteractor: &moduleInteractor,
				moduleIdentifier: "sample-module",
				channel:          "",
				kymaChannel:      "regular",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateChannel(tt.args.ctx, tt.args.moduleInteractor, tt.args.moduleIdentifier, tt.args.channel,
				tt.args.kymaChannel); (err != nil) != tt.wantErr {
				t.Errorf("validateChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
