package diagnostics

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/stretchr/testify/assert/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Data returned by the collector
type Metadata struct {
	GlobalAccountID   string
	SubaccountID      string
	Provider          string
	KubernetesVersion string
	NATGatewayIPs     []string
	KubeAPIServer     string
}

type MetadataCollector struct {
	client kube.Client
	VerboseLogger
}

func NewMetadataCollector(client kube.Client, writer io.Writer, verbose bool) *MetadataCollector {
	return &MetadataCollector{
		client:        client,
		VerboseLogger: NewVerboseLogger(writer, verbose),
	}
}

func (mc *MetadataCollector) Run(ctx context.Context) Metadata {
	var metadata Metadata

	mc.enrichMetadataWithShootInfo(ctx, &metadata)
	mc.enrichMetadataWithKymaInfo(ctx, &metadata)
	mc.enrichMetadataWithKymaProvisioningInfo(ctx, &metadata)
	mc.enrichMetadataWithClusterConfigInfo(ctx, &metadata)

	return metadata
}

func (mc *MetadataCollector) enrichMetadataWithClusterConfigInfo(ctx context.Context, metadata *Metadata) {
	networkProblemDetectorConfigMap, err := mc.client.Static().CoreV1().
		ConfigMaps("kube-system").
		Get(ctx, "network-problem-detector-cluster-config", metav1.GetOptions{})

	if err != nil {
		mc.WriteVerboseError(err, "Failed to get network-problem-detector-cluster-config ConfigMap from kube-system namespace")
		return
	}

	if clusterConfigYAML, exists := networkProblemDetectorConfigMap.Data["cluster-config.yaml"]; exists {
		type KubeAPIServer struct {
			Hostname string `yaml:"hostname"`
		}

		type ClusterConfig struct {
			KubeAPIServer KubeAPIServer `yaml:"kubeAPIServer"`
		}

		var cc ClusterConfig
		err := yaml.Unmarshal([]byte(clusterConfigYAML), &cc)
		if err != nil {
			mc.WriteVerboseError(err, "Failed to unmarshal provisioning details YAML")
			return
		}

		metadata.KubeAPIServer = fmt.Sprintf("https://%s", cc.KubeAPIServer.Hostname)
	}
}

func (mc *MetadataCollector) enrichMetadataWithShootInfo(ctx context.Context, metadata *Metadata) {
	shootInfoConfigMap, err := mc.client.Static().CoreV1().
		ConfigMaps("kube-system").
		Get(ctx, "shoot-info", metav1.GetOptions{})

	if err != nil {
		mc.WriteVerboseError(err, "Failed to get shoot-info ConfigMap from kube-system namespace")
		return
	}

	if provider, exists := shootInfoConfigMap.Data["provider"]; exists {
		metadata.Provider = provider
	}

	if k8sVersion, exists := shootInfoConfigMap.Data["kubernetesVersion"]; exists {
		metadata.KubernetesVersion = k8sVersion
	}
}

func (mc *MetadataCollector) enrichMetadataWithKymaInfo(ctx context.Context, metadata *Metadata) {
	kymaInfoConfigMap, err := mc.client.Static().CoreV1().
		ConfigMaps("kyma-system").
		Get(ctx, "kyma-info", metav1.GetOptions{})

	if err != nil {
		mc.WriteVerboseError(err, "Failed to get kyma-info ConfigMap from kyma-system namespace")
		return
	}

	if natGatewayIPs, exists := kymaInfoConfigMap.Data["cloud.natGatewayIps"]; exists {
		splitGatewayIPs := strings.Split(natGatewayIPs, " ")
		metadata.NATGatewayIPs = append(metadata.NATGatewayIPs, splitGatewayIPs...)
	}
}

func (mc *MetadataCollector) enrichMetadataWithKymaProvisioningInfo(ctx context.Context, metadata *Metadata) {
	kymaProvisioningConfigMap, err := mc.client.Static().CoreV1().
		ConfigMaps("kyma-system").
		Get(ctx, "kyma-provisioning-info", metav1.GetOptions{})

	if err != nil {
		mc.WriteVerboseError(err, "Failed to get kyma-provisioning-info ConfigMap from kyma-system namespace")
		return
	}

	if detailsYAML, exists := kymaProvisioningConfigMap.Data["details"]; exists {
		type ProvisioningDetails struct {
			GlobalAccountID string `yaml:"globalAccountID"`
			SubaccountID    string `yaml:"subaccountID"`
		}

		var details ProvisioningDetails
		err := yaml.Unmarshal([]byte(detailsYAML), &details)
		if err != nil {
			mc.WriteVerboseError(err, "Failed to unmarshal provisioning details YAML")
			return
		}

		metadata.GlobalAccountID = details.GlobalAccountID
		metadata.SubaccountID = details.SubaccountID
	}
}
