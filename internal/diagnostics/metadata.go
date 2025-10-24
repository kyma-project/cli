package diagnostics

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/assert/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Metadata struct {
	GlobalAccountID    string   `json:"globalAccountID" yaml:"globalAccountID"`
	SubaccountID       string   `json:"subaccountID" yaml:"subaccountID"`
	ClusterID          string   `json:"clusterID" yaml:"clusterID"`
	ClusterDomain      string   `json:"clusterDomain" yaml:"clusterDomain"`
	Region             string   `json:"region" yaml:"region"`
	ShootName          string   `json:"shootName" yaml:"shootName"`
	Provider           string   `json:"provider" yaml:"provider"`
	KubernetesVersion  string   `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	NATGatewayIPs      []string `json:"natGatewayIPs" yaml:"natGatewayIPs"`
	GardenerExtensions []string `json:"gardenerExtensions" yaml:"gardenerExtensions"`
	KubeAPIServer      string   `json:"kubeAPIServer" yaml:"kubeAPIServer"`
}

type MetadataCollector struct {
	client kube.Client
	*out.Printer
}

func NewMetadataCollector(client kube.Client) *MetadataCollector {
	return &MetadataCollector{
		client:  client,
		Printer: out.Default,
	}
}

func (mc *MetadataCollector) Run(ctx context.Context) Metadata {
	var metadata Metadata

	mc.enrichMetadataWithSapBtpManagerSecret(ctx, &metadata)
	mc.enrichMetadataWithShootInfo(ctx, &metadata)
	mc.enrichMetadataWithKymaInfo(ctx, &metadata)
	mc.enrichMetadataWithKymaProvisioningInfo(ctx, &metadata)
	mc.enrichMetadataWithClusterConfigInfo(ctx, &metadata)

	return metadata
}

func (mc *MetadataCollector) enrichMetadataWithSapBtpManagerSecret(ctx context.Context, metadata *Metadata) {
	secret, err := mc.client.Static().CoreV1().Secrets("kyma-system").Get(ctx, "sap-btp-manager", metav1.GetOptions{})
	if err != nil {
		mc.Verbosefln("Failed to get sap-btp-manager secret: %v", err)
		return
	}

	if secret.Data["cluster_id"] == nil {
		return
	}

	metadata.ClusterID = string(secret.Data["cluster_id"])
}

func (mc *MetadataCollector) enrichMetadataWithClusterConfigInfo(ctx context.Context, metadata *Metadata) {
	networkProblemDetectorConfigMap, err := mc.client.Static().CoreV1().
		ConfigMaps("kube-system").
		Get(ctx, "network-problem-detector-cluster-config", metav1.GetOptions{})

	if err != nil {
		mc.Verbosefln("Failed to get network-problem-detector-cluster-config ConfigMap from kube-system namespace: %v", err)
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
			mc.Verbosefln("Failed to unmarshal provisioning details YAML: %v", err)
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
		mc.Verbosefln("Failed to get shoot-info ConfigMap from kube-system namespace: %v", err)
		return
	}

	if provider, exists := shootInfoConfigMap.Data["provider"]; exists {
		metadata.Provider = provider
	}

	if k8sVersion, exists := shootInfoConfigMap.Data["kubernetesVersion"]; exists {
		metadata.KubernetesVersion = k8sVersion
	}

	if domain, exists := shootInfoConfigMap.Data["domain"]; exists {
		metadata.ClusterDomain = domain
	}

	if region, exists := shootInfoConfigMap.Data["region"]; exists {
		metadata.Region = region
	}

	if shootName, exists := shootInfoConfigMap.Data["shootName"]; exists {
		metadata.ShootName = shootName
	}

	if extensions, exists := shootInfoConfigMap.Data["extensions"]; exists {
		splitExtensions := strings.Split(extensions, ",")
		metadata.GardenerExtensions = append(metadata.GardenerExtensions, splitExtensions...)
	}
}

func (mc *MetadataCollector) enrichMetadataWithKymaInfo(ctx context.Context, metadata *Metadata) {
	kymaInfoConfigMap, err := mc.client.Static().CoreV1().
		ConfigMaps("kyma-system").
		Get(ctx, "kyma-info", metav1.GetOptions{})

	if err != nil {
		mc.Verbosefln("Failed to get kyma-info ConfigMap from kyma-system namespace: %v", err)
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
		mc.Verbosefln("Failed to get kyma-provisioning-info ConfigMap from kyma-system namespace: %v", err)
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
			mc.Verbosefln("Failed to unmarshal provisioning details YAML: %v", err)
			return
		}

		metadata.GlobalAccountID = details.GlobalAccountID
		metadata.SubaccountID = details.SubaccountID
	}
}
