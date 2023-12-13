package deploy

import (
	"context"
	"testing"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube/mocks"
	"github.com/kyma-project/cli/pkg/step"
)

func Test_command_detectManagedKyma(t *testing.T) {
	type fields struct {
		Command cli.Command
		opts    *Options
	}
	type args struct {
		ctx context.Context
	}

	// Managed Kyma
	managedKymaMock := &mocks.KymaKube{}
	managedKyma := &unstructured.Unstructured{}
	managedKyma.SetKind(string(shared.KymaKind))
	managedKyma.SetAPIVersion("operator.kyma-project.io/v1beta2")
	managedKyma.SetManagedFields([]metav1.ManagedFieldsEntry{
		{
			Manager:     "lifecycle-manager",
			Subresource: "status",
		},
	})
	managedDynamic := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, managedKyma)
	managedKymaMock.On("Dynamic").Return(managedDynamic)

	// Unmanaged Kyma
	unmanagedKymaMock := &mocks.KymaKube{}
	unmanagedKyma := &unstructured.Unstructured{}
	unmanagedKyma.SetKind(string(shared.KymaKind))
	unmanagedKyma.SetAPIVersion("operator.kyma-project.io/v1beta2")
	unmanagedKyma.SetManagedFields([]metav1.ManagedFieldsEntry{
		{
			Manager:     "unmanaged-kyma",
			Subresource: "status",
		},
	})
	unmanagedDynamic := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, unmanagedKyma)
	unmanagedKymaMock.On("Dynamic").Return(unmanagedDynamic)

	// kyma with no managed fields
	noManagedFieldsMock := &mocks.KymaKube{}
	noManagedFields := &unstructured.Unstructured{}
	noManagedFields.SetKind(string(shared.KymaKind))
	noManagedFields.SetAPIVersion("operator.kyma-project.io/v1beta2")
	noManagedFieldsDynamic := fakedynamic.NewSimpleDynamicClient(scheme.Scheme, noManagedFields)
	noManagedFieldsMock.On("Dynamic").Return(noManagedFieldsDynamic)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "managed kyma",
			fields: fields{
				Command: cli.Command{
					K8s:         managedKymaMock,
					CurrentStep: step.NewMutedStep(),
				},
				opts: &Options{Options: cli.NewOptions()},
			},
			wantErr: true,
			args: args{
				ctx: context.TODO(),
			},
		},
		{
			name: "unmanaged kyma",
			fields: fields{
				Command: cli.Command{
					K8s:         unmanagedKymaMock,
					CurrentStep: step.NewMutedStep(),
				},
				opts: &Options{Options: cli.NewOptions()},
			},
			wantErr: false,
			args: args{
				ctx: context.TODO(),
			},
		},
		{
			name: "no kyma found",
			fields: fields{
				Command: cli.Command{
					K8s:         noManagedFieldsMock,
					CurrentStep: step.NewMutedStep(),
				},
				opts: &Options{Options: cli.NewOptions()},
			},
			wantErr: false,
			args: args{
				ctx: context.TODO(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command{
				Command: tt.fields.Command,
				opts:    tt.fields.opts,
			}
			if err := cmd.detectManagedKyma(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("detectManagedKyma() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
