package diagnostics

import (
	"bytes"
	"context"
	"strings"
	"testing"

	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/out"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewMetadataCollector(t *testing.T) {
	// Given
	kubeClient := &kube_fake.KubeClient{
		TestKubernetesInterface: fake.NewSimpleClientset(),
	}

	// When
	collector := NewMetadataCollector(kubeClient)

	// Then
	assert.NotNil(t, collector)
}

func TestEnrichMetadataWithShootInfo(t *testing.T) {
	// Test cases
	testCases := []struct {
		name               string
		shootInfoExists    bool
		shootInfoConfigMap *corev1.ConfigMap
		verbose            bool
	}{
		{
			name:            "Should enrich metadata when shoot-info ConfigMap exists",
			shootInfoExists: true,
			shootInfoConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shoot-info",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"provider":          "azure",
					"kubernetesVersion": "1.26.0",
					"domain":            "c-513b462.sample-domain.com",
					"region":            "westeurope",
					"shootName":         "c-513b462",
					"extensions":        "extension-1,extension-2,extension-3",
				},
			},
			verbose: false,
		},
		{
			name:               "Should not enrich metadata when shoot-info ConfigMap does not exist",
			shootInfoExists:    false,
			shootInfoConfigMap: &corev1.ConfigMap{},
			verbose:            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			fakeClient := fake.NewSimpleClientset()
			kubeClient := &kube_fake.KubeClient{
				TestKubernetesInterface: fakeClient,
			}

			// Create the ConfigMap if needed for the test case
			if tc.shootInfoExists {
				_, err := fakeClient.
					CoreV1().
					ConfigMaps("kube-system").
					Create(context.TODO(), tc.shootInfoConfigMap, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			printer := out.NewToWriter(&writer)
			if tc.verbose {
				printer.EnableVerbose()
			}

			// When
			collector := MetadataCollector{kubeClient, printer}
			metadata := collector.Run(context.TODO())

			// Then
			if tc.shootInfoExists {
				assert.Equal(t, tc.shootInfoConfigMap.Data["provider"], metadata.Provider)
				assert.Equal(t, tc.shootInfoConfigMap.Data["kubernetesVersion"], metadata.KubernetesVersion)
				assert.Equal(t, tc.shootInfoConfigMap.Data["domain"], metadata.ClusterDomain)
				assert.Equal(t, tc.shootInfoConfigMap.Data["region"], metadata.Region)
				assert.Equal(t, tc.shootInfoConfigMap.Data["shootName"], metadata.ShootName)
				assert.Equal(t, tc.shootInfoConfigMap.Data["extensions"], strings.Join(metadata.GardenerExtensions, ","))
			} else {
				assert.Empty(t, metadata.Provider)
				assert.Empty(t, metadata.KubernetesVersion)
				if tc.verbose {
					assert.Contains(t, writer.String(), "Failed to get shoot-info ConfigMap")
				} else {
					assert.Empty(t, writer.String())
				}
			}
		})
	}
}

func TestEnrichMetadataWithKymaInfo(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		kymaInfoExists bool
		natGatewayIPs  string
		expectedIPs    []string
		verbose        bool
	}{
		{
			name:           "Should enrich metadata when kyma-info ConfigMap exists",
			kymaInfoExists: true,
			natGatewayIPs:  "192.168.0.1 192.168.0.2",
			expectedIPs:    []string{"192.168.0.1", "192.168.0.2"},
			verbose:        false,
		},
		{
			name:           "Should not enrich metadata when kyma-info ConfigMap does not exist",
			kymaInfoExists: false,
			natGatewayIPs:  "",
			expectedIPs:    nil,
			verbose:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			fakeClient := fake.NewSimpleClientset()
			kubeClient := &kube_fake.KubeClient{
				TestKubernetesInterface: fakeClient,
			}

			// Create the ConfigMap if needed for the test case
			if tc.kymaInfoExists {
				kymaInfoCM := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kyma-info",
						Namespace: "kyma-system",
					},
					Data: map[string]string{
						"cloud.natGatewayIps": tc.natGatewayIPs,
					},
				}
				_, err := fakeClient.CoreV1().ConfigMaps("kyma-system").Create(context.TODO(), kymaInfoCM, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			printer := out.NewToWriter(&writer)
			if tc.verbose {
				printer.EnableVerbose()
			}

			// When
			collector := MetadataCollector{kubeClient, printer}
			metadata := collector.Run(context.TODO())

			// Then
			if tc.kymaInfoExists {
				assert.Equal(t, tc.expectedIPs, metadata.NATGatewayIPs)
			} else {
				assert.Empty(t, metadata.NATGatewayIPs)
				if tc.verbose {
					assert.Contains(t, writer.String(), "Failed to get kyma-info ConfigMap")
				} else {
					assert.Empty(t, writer.String())
				}
			}
		})
	}
}

func TestEnrichMetadataWithKymaProvisioningInfo(t *testing.T) {
	// Test cases
	testCases := []struct {
		name                    string
		provisioningInfoExists  bool
		detailsYAML             string
		expectedGlobalAccountID string
		expectedSubaccountID    string
		verbose                 bool
	}{
		{
			name:                    "Should enrich metadata when kyma-provisioning-info ConfigMap exists with valid YAML",
			provisioningInfoExists:  true,
			detailsYAML:             "globalAccountID: ga-12345\nsubaccountID: sa-67890",
			expectedGlobalAccountID: "ga-12345",
			expectedSubaccountID:    "sa-67890",
			verbose:                 false,
		},
		{
			name:                    "Should not enrich metadata when kyma-provisioning-info ConfigMap does not exist",
			provisioningInfoExists:  false,
			detailsYAML:             "",
			expectedGlobalAccountID: "",
			expectedSubaccountID:    "",
			verbose:                 true,
		},
		{
			name:                    "Should handle invalid YAML in details field",
			provisioningInfoExists:  true,
			detailsYAML:             "invalid: yaml: :",
			expectedGlobalAccountID: "",
			expectedSubaccountID:    "",
			verbose:                 true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			fakeClient := fake.NewSimpleClientset()
			kubeClient := &kube_fake.KubeClient{
				TestKubernetesInterface: fakeClient,
			}

			// Create the ConfigMap if needed for the test case
			if tc.provisioningInfoExists {
				provisioningInfoCM := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kyma-provisioning-info",
						Namespace: "kyma-system",
					},
					Data: map[string]string{
						"details": tc.detailsYAML,
					},
				}
				_, err := fakeClient.CoreV1().ConfigMaps("kyma-system").Create(context.TODO(), provisioningInfoCM, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			printer := out.NewToWriter(&writer)
			if tc.verbose {
				printer.EnableVerbose()
			}

			// When
			collector := MetadataCollector{kubeClient, printer}
			metadata := collector.Run(context.TODO())

			// Then
			assert.Equal(t, tc.expectedGlobalAccountID, metadata.GlobalAccountID)
			assert.Equal(t, tc.expectedSubaccountID, metadata.SubaccountID)

			if !tc.provisioningInfoExists && tc.verbose {
				assert.Contains(t, writer.String(), "Failed to get kyma-provisioning-info ConfigMap")
			} else if tc.provisioningInfoExists && tc.detailsYAML == "invalid: yaml: :" && tc.verbose {
				assert.Contains(t, writer.String(), "Failed to unmarshal provisioning details YAML")
			}
		})
	}
}

func TestEnrichMetadataWithSapBtpManagerSecret(t *testing.T) {
	testCases := []struct {
		name                string
		secretExists        bool
		secret              *corev1.Secret
		expectedClusterID   string
		verbose             bool
		expectedErrorOutput string
	}{
		{
			name:         "Should enrich metadata when sap-btp-manager secret exists with cluster_id",
			secretExists: true,
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sap-btp-manager",
					Namespace: "kyma-system",
				},
				Data: map[string][]byte{
					"cluster_id": []byte("test-cluster-123"),
				},
			},
			expectedClusterID: "test-cluster-123",
			verbose:           false,
		},
		{
			name:         "Should not enrich metadata when sap-btp-manager secret exists but cluster_id is missing",
			secretExists: true,
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sap-btp-manager",
					Namespace: "kyma-system",
				},
				Data: map[string][]byte{
					"other_data": []byte("some-value"),
				},
			},
			expectedClusterID: "",
			verbose:           false,
		},
		{
			name:                "Should handle missing sap-btp-manager secret gracefully",
			secretExists:        false,
			secret:              nil,
			expectedClusterID:   "",
			verbose:             true,
			expectedErrorOutput: "Failed to get sap-btp-manager secret",
		},
		{
			name:         "Should handle empty cluster_id data",
			secretExists: true,
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sap-btp-manager",
					Namespace: "kyma-system",
				},
				Data: map[string][]byte{
					"cluster_id": []byte(""),
				},
			},
			expectedClusterID: "",
			verbose:           false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			fakeClient := fake.NewSimpleClientset()
			kubeClient := &kube_fake.KubeClient{
				TestKubernetesInterface: fakeClient,
			}

			if tc.secretExists {
				_, err := fakeClient.CoreV1().Secrets("kyma-system").Create(context.TODO(), tc.secret, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			printer := out.NewToWriter(&writer)
			if tc.verbose {
				printer.EnableVerbose()
			}

			// When
			collector := MetadataCollector{kubeClient, printer}
			metadata := collector.Run(context.TODO())

			// Then
			assert.Equal(t, tc.expectedClusterID, metadata.ClusterID)

			if tc.expectedErrorOutput != "" && tc.verbose {
				assert.Contains(t, writer.String(), tc.expectedErrorOutput)
			}
		})
	}
}
