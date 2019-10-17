package gardener

import (
	"testing"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestProvisionGardenerFlags ensures that the provided command flags are stored in the options.
func TestProvisionGardenerFlags(t *testing.T) {
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, "", o.Name, "Default value for the name flag not as expected.")
	require.Equal(t, "", o.Project, "Default value for the project flag not as expected.")
	require.Equal(t, "", o.CredentialsFile, "Default value for the credentials flag not as expected.")
	require.Equal(t, "gcp", o.TargetProvider, "The parsed value for the credentials flag should gcp")
	require.Equal(t, "", o.Secret, "The parsed value for the secret flag not as expected.")
	require.Equal(t, "1.15.4", o.KubernetesVersion, "Default value for the kube-version flag not as expected.")
	require.Equal(t, "europe-west3", o.Region, "Default value for the region flag not as expected.")
	require.Equal(t, "europe-west3-a", o.Zone, "Default value for the zone flag not as expected.")
	require.Equal(t, "n1-standard-4", o.MachineType, "Default value for the type flag not as expected.")
	require.Equal(t, 30, o.DiskSizeGB, "Default value for the disk-size flag not as expected.")
	require.Equal(t, "pd-standard", o.DiskType, "Default value for the disk-type flag not as expected.")
	require.Equal(t, 2, o.NodeCount, "Default value for the nodes flag not as expected.")
	require.Equal(t, 2, o.ScalerMin, "Default value for the scaler-min flag not as expected.")
	require.Equal(t, 4, o.ScalerMax, "Default value for the scaler-max flag not as expected.")
	require.Equal(t, 4, o.Surge, "Default value for the surge flag not as expected.")
	require.Equal(t, 1, o.Unavailable, "Default value for the unavailable flag not as expected.")
	require.Equal(t, "10.250.0.0/19", o.CIDR, "Default value for the cidr flag not as expected.")
	require.Empty(t, o.Extra, "Default value for the extra flag not as expected.")

	// test passing flags
	c.ParseFlags([]string{
		"-n", "my-cluster",
		"-p", "my-project",
		"-c", "/my/credentials/file",
		"--target-provider", "Alibaba",
		"-s", "my-ali-key",
		"--disk-type", "a big one",
		"-k", "1.16.0",
		"-r", "us-central",
		"-z", "us-central1-b",
		"-t", "quantum-computer",
		"--disk-size", "2000",
		"--nodes", "7",
		"--scaler-min", "88",
		"--scaler-max", "99",
		"--surge", "100",
		"-u", "45",
		"--cidr", "0.0.0.0/24",
		"--extra", "VAR1=VALUE1,VAR2=VALUE2",
	})

	require.Equal(t, "my-cluster", o.Name, "The parsed value for the name flag not as expected.")
	require.Equal(t, "my-project", o.Project, "The parsed value for the project flag not as expected.")
	require.Equal(t, "/my/credentials/file", o.CredentialsFile, "The parsed value for the credentials flag not as expected.")
	require.Equal(t, "Alibaba", o.TargetProvider, "The parsed value for the target-provider flag not as expected.")
	require.Equal(t, "my-ali-key", o.Secret, "The parsed value for the secret flag not as expected.")
	require.Equal(t, "1.16.0", o.KubernetesVersion, "The parsed value for the kube-version flag not as expected.")
	require.Equal(t, "us-central", o.Region, "The parsed value for the region flag not as expected.")
	require.Equal(t, "us-central1-b", o.Zone, "The parsed value for the zone flag not as expected.")
	require.Equal(t, "quantum-computer", o.MachineType, "The parsed value for the type flag not as expected.")
	require.Equal(t, 2000, o.DiskSizeGB, "The parsed value for the disk-size flag not as expected.")
	require.Equal(t, "a big one", o.DiskType, "The parsed value for the disk-type flag not as expected.")
	require.Equal(t, 7, o.NodeCount, "The parsed value for the nodes flag not as expected.")
	require.Equal(t, 88, o.ScalerMin, "The parsed value for the scaler-min flag not as expected.")
	require.Equal(t, 99, o.ScalerMax, "The parsed value for the scaler-max flag not as expected.")
	require.Equal(t, 100, o.Surge, "The parsed value for the surge flag not as expected.")
	require.Equal(t, 45, o.Unavailable, "The parsed value for the unavailable flag not as expected.")
	require.Equal(t, "0.0.0.0/24", o.CIDR, "The parsed value for the cidr flag not as expected.")
	require.Equal(t, []string{"VAR1=VALUE1", "VAR2=VALUE2"}, o.Extra, "The parsed value for the extra flag not as expected.")
}

func TestProvisionGardenerSubcommands(t *testing.T) {
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
		NodeCount:         3,
	}

	c := newCluster(o)
	require.Equal(t, o.Name, c.Name, "Cluster name not as expected.")
	require.Equal(t, o.KubernetesVersion, c.KubernetesVersion, "Cluster Kubernetes version not as expected.")
	require.Equal(t, o.Region, c.Location, "Cluster location not as expected.")
	require.Equal(t, o.MachineType, c.MachineType, "Cluster machine type not as expected.")
	require.Equal(t, o.DiskSizeGB, c.DiskSizeGB, "Cluster disk size not as expected.")
	require.Equal(t, o.NodeCount, c.NodeCount, "Cluster number of nodes not as expected.")
}

func TestNewProvider(t *testing.T) {
	o := &Options{
		Project:         "cool-project",
		CredentialsFile: "/path/to/credentials",
		TargetProvider:  "AlibabaCloud",
		Secret:          "Open sesame!",
		Zone:            "Desert",
		DiskType:        "a big one",
		CIDR:            "0.0.0.0/24",
		ScalerMin:       12,
		ScalerMax:       26,
		Surge:           35,
		Unavailable:     5,
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
	custom["target_provider"] = o.TargetProvider
	custom["zone"] = o.Zone
	custom["disk_type"] = o.DiskType
	custom["autoscaler_min"] = o.ScalerMin
	custom["autoscaler_max"] = o.ScalerMax
	custom["max_surge"] = o.Surge
	custom["max_unavailable"] = o.Unavailable
	custom["cidr"] = o.CIDR
	require.Equal(t, custom, p.CustomConfigurations, "Provider extra configurations not as expected.")
}
