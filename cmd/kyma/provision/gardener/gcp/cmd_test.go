package gcp

import (
	"testing"

	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestProvisionGardenerFlags ensures that the provided command flags are stored in the options.
func TestProvisionGardenerGCPFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, "", o.Name, "Default value for the name flag not as expected.")
	require.Equal(t, "", o.Project, "Default value for the project flag not as expected.")
	require.Equal(t, "", o.CredentialsFile, "Default value for the credentials flag not as expected.")
	require.Equal(t, "", o.Secret, "The parsed value for the secret flag not as expected.")
	require.Equal(t, "1.16", o.KubernetesVersion, "Default value for the kube-version flag not as expected.")
	require.Equal(t, "europe-west3", o.Region, "Default value for the region flag not as expected.")
	require.Equal(t, []string{"europe-west3-a"}, o.Zones, "Default value for the zone flag not as expected.")
	require.Equal(t, "n1-standard-4", o.MachineType, "Default value for the type flag not as expected.")
	require.Equal(t, 50, o.DiskSizeGB, "Default value for the disk-size flag not as expected.")
	require.Equal(t, "pd-standard", o.DiskType, "Default value for the disk-type flag not as expected.")
	require.Equal(t, 2, o.ScalerMin, "Default value for the scaler-min flag not as expected.")
	require.Equal(t, 3, o.ScalerMax, "Default value for the scaler-max flag not as expected.")
	require.Empty(t, o.Extra, "Default value for the extra flag not as expected.")
	require.Equal(t, uint(3), o.Attempts, "Default value for the attempts flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"-n", "my-cluster",
		"-p", "my-project",
		"-c", "/my/credentials/file",
		"-s", "my-ali-key",
		"--disk-type", "a big one",
		"-k", "1.16.0",
		"-r", "us-central",
		"-z", "us-central1-b",
		"-t", "quantum-computer",
		"--disk-size", "2000",
		"--scaler-min", "88",
		"--scaler-max", "99",
		"--extra", "VAR1=VALUE1,VAR2=VALUE2",
		"--attempts", "2",
	})

	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "my-cluster", o.Name, "The parsed value for the name flag not as expected.")
	require.Equal(t, "my-project", o.Project, "The parsed value for the project flag not as expected.")
	require.Equal(t, "/my/credentials/file", o.CredentialsFile, "The parsed value for the credentials flag not as expected.")
	require.Equal(t, "my-ali-key", o.Secret, "The parsed value for the secret flag not as expected.")
	require.Equal(t, "1.16.0", o.KubernetesVersion, "The parsed value for the kube-version flag not as expected.")
	require.Equal(t, "us-central", o.Region, "The parsed value for the region flag not as expected.")
	require.Equal(t, []string{"us-central1-b"}, o.Zones, "The parsed value for the zone flag not as expected.")
	require.Equal(t, "quantum-computer", o.MachineType, "The parsed value for the type flag not as expected.")
	require.Equal(t, 2000, o.DiskSizeGB, "The parsed value for the disk-size flag not as expected.")
	require.Equal(t, "a big one", o.DiskType, "The parsed value for the disk-type flag not as expected.")
	require.Equal(t, 88, o.ScalerMin, "The parsed value for the scaler-min flag not as expected.")
	require.Equal(t, 99, o.ScalerMax, "The parsed value for the scaler-max flag not as expected.")
	require.Equal(t, []string{"VAR1=VALUE1", "VAR2=VALUE2"}, o.Extra, "The parsed value for the extra flag not as expected.")
	require.Equal(t, uint(2), o.Attempts, "The parsed value for the attempts flag not as expected.")
}

func TestProvisionGardenerGCPSubcommands(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	sub := c.Commands()

	require.Equal(t, 0, len(sub), "Number of provision gardener subcommands not as expected")
}

func TestNewCluster(t *testing.T) {
	o := &Options{
		Name:              "mega-cluster",
		KubernetesVersion: "1.16.0",
		Region:            "north-pole",
		MachineType:       "HAL",
		DiskSizeGB:        9000,
		ScalerMax:         3,
	}

	c := newCluster(o)
	require.Equal(t, o.Name, c.Name, "Cluster name not as expected.")
	require.Equal(t, o.KubernetesVersion, c.KubernetesVersion, "Cluster Kubernetes version not as expected.")
	require.Equal(t, o.Region, c.Location, "Cluster location not as expected.")
	require.Equal(t, o.MachineType, c.MachineType, "Cluster machine type not as expected.")
	require.Equal(t, o.DiskSizeGB, c.DiskSizeGB, "Cluster disk size not as expected.")
	require.Equal(t, o.ScalerMax, c.NodeCount, "Cluster number of nodes not as expected.")
}

func TestNewProvider(t *testing.T) {
	t.Parallel()
	o := &Options{
		Project:         "cool-project",
		CredentialsFile: "/path/to/credentials",
		Secret:          "Open sesame!",
		Zones:           []string{"Desert"},
		DiskType:        "a big one",
		ScalerMin:       12,
		ScalerMax:       26,
		Extra:           []string{"VAR1=VALUE1", "VAR2=VALUE2"},
	}

	p, err := newProvider(o)
	require.NoError(t, err)

	require.Equal(t, types.Gardener, p.Type, "Provider type not as expected.")
	require.Equal(t, o.Project, p.ProjectName, "Provider project name not as expected.")
	require.Equal(t, o.CredentialsFile, p.CredentialsFilePath, "Provider credentials file path not as expected.")

	custom := make(map[string]interface{})
	custom["VAR1"] = "VALUE1"
	custom["VAR2"] = "VALUE2"
	custom["target_secret"] = o.Secret
	custom["target_provider"] = "gcp"
	custom["gcp_control_plane_zone"] = o.Zones[0]
	custom["zones"] = o.Zones
	custom["disk_type"] = o.DiskType
	custom["worker_minimum"] = o.ScalerMin
	custom["worker_maximum"] = o.ScalerMax
	custom["worker_max_surge"] = 1
	custom["worker_max_unavailable"] = 1
	custom["vnetcidr"] = "10.250.0.0/16"
	custom["workercidr"] = "10.250.0.0/16"
	custom["networking_type"] = "calico"
	custom["machine_image_name"] = "gardenlinux"
	custom["machine_image_version"] = "27.1.0"

	require.Equal(t, custom, p.CustomConfigurations, "Provider extra configurations not as expected.")
}
