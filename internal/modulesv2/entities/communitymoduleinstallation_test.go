package entities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommunityModuleInstallation_FullName_ReturnsNamespacedName(t *testing.T) {
	m := CommunityModuleInstallation{Name: "docker-registry", Namespace: "default"}
	require.Equal(t, "default/docker-registry", m.FullName())
}
