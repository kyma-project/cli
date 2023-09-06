package module

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-project/cli/cmd/kyma/alpha/enable/module/mock"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"github.com/kyma-project/lifecycle-manager/pkg/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestChannelValidation(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	template1, _ := testutils.ModuleTemplateFactory(v1beta2.Module{
		Name:                 "test",
		ControllerName:       "-",
		Channel:              "fast",
		CustomResourcePolicy: "-",
	}, unstructured.Unstructured{}, false, false, false, false)
	template2, _ := testutils.ModuleTemplateFactory(v1beta2.Module{
		Name:                 "not-test",
		ControllerName:       "-",
		Channel:              "alpha",
		CustomResourcePolicy: "-",
	}, unstructured.Unstructured{}, false, false, false, false)
	allTemplates := v1beta2.ModuleTemplateList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []v1beta2.ModuleTemplate{
			*template1, *template2,
		},
	}

	moduleInteractor := mock.Interactor{}
	moduleInteractor.Test(t)
	moduleInteractor.On("GetAllModuleTemplates", ctx).Return(allTemplates, nil)

	// WHEN 1
	moduleName := "test"
	channel := "alpha"
	kymaChannel := "regular"
	err := validateChannel(ctx, &moduleInteractor, moduleName, channel, kymaChannel)
	// THEN 1
	if !strings.Contains(err.Error(), "the channel ["+channel+"] does not exist") {
		t.Fatalf("channel validation failed. invalid channel [%s] did not throw error.", channel)
	}

	// WHEN 2
	channel = "fast"
	err = validateChannel(ctx, &moduleInteractor, moduleName, channel, kymaChannel)
	// THEN 2
	if err != nil && strings.Contains(err.Error(), "the channel ["+channel+"] does not exist") {
		t.Fatalf("channel validation failed. valid channel [%s] threw an error.", channel)
	}
}
