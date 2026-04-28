package entities

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/stretchr/testify/require"
)

func TestModuleInstallation_IsManaged_TrueWhenManagedIsNil(t *testing.T) {
	m := ModuleInstallation{Managed: nil}
	require.True(t, m.IsManaged())
}

func TestModuleInstallation_IsManaged_TrueWhenManagedIsTrue(t *testing.T) {
	managed := true
	m := ModuleInstallation{Managed: &managed}
	require.True(t, m.IsManaged())
}

func TestModuleInstallation_IsManaged_FalseWhenManagedIsFalse(t *testing.T) {
	managed := false
	m := ModuleInstallation{Managed: &managed}
	require.False(t, m.IsManaged())
}

func TestNewModuleInstallationFromRaw_MapsNameVersionChannel(t *testing.T) {
	raw := kyma.KymaModuleInfo{
		Status: kyma.ModuleStatus{Name: "api-gateway", Version: "3.5.1", Channel: "regular"},
	}

	m := NewModuleInstallationFromRaw(raw)

	require.Equal(t, "api-gateway", m.Name)
	require.Equal(t, "3.5.1", m.Version)
	require.Equal(t, "regular", m.Channel)
}

func TestNewModuleInstallationFromRaw_MapsModuleState(t *testing.T) {
	raw := kyma.KymaModuleInfo{
		Status: kyma.ModuleStatus{Name: "api-gateway", State: "Ready"},
	}

	m := NewModuleInstallationFromRaw(raw)

	require.Equal(t, "Ready", m.ModuleState)
}

func TestNewModuleInstallationFromRaw_MapsManaged(t *testing.T) {
	managed := false
	raw := kyma.KymaModuleInfo{
		Spec:   kyma.Module{Managed: &managed},
		Status: kyma.ModuleStatus{Name: "api-gateway"},
	}

	m := NewModuleInstallationFromRaw(raw)

	require.NotNil(t, m.Managed)
	require.False(t, *m.Managed)
}

func TestNewModuleInstallationFromRaw_MapsCustomResourcePolicy(t *testing.T) {
	raw := kyma.KymaModuleInfo{
		Spec:   kyma.Module{CustomResourcePolicy: "CreateAndDelete"},
		Status: kyma.ModuleStatus{Name: "api-gateway"},
	}

	m := NewModuleInstallationFromRaw(raw)

	require.Equal(t, "CreateAndDelete", m.CustomResourcePolicy)
}
