package installation

import "time"

// Options holds the configuration options.
type Options struct {
	// Source specifies the installation source. To use the specific release, pass the release version (e.g. 1.6.0).
	// To use the latest master, pass "latest". To use the local sources, pass "local". To use the remote image, pass the installer image (e.g. user/my-kyma-installer:v1.6.0).
	Source string `json:"source"`

	// releaseVersion is set to the version of the release being installed.
	releaseVersion string
	// configVersion is set to the version of the configuration files being used.
	configVersion string
	// remoteImage holds the image URL if the installation source is an image.
	remoteImage string
	// registryTemplate specifies the registry image pattern.
	registryTemplate string
	// fromLocalSources is set if the installation source is local.
	fromLocalSources bool

	// LocalSrcPath specifies the absolute path to local sources.
	LocalSrcPath string `json:"localSrcPath"`
	// OverrideConfigs specifies the path to a yaml file with parameters to override.
	OverrideConfigs []string `json:"overrideConfigs"`
	// Password specifies the predefined cluster password.
	Password string `json:"password"`
	// Domain specifies the domain used for installation.
	Domain string `json:"domain"`
	// TLSCert specifies the TLS certificate for the domain used for installation
	TLSCert string `json:"tlsCert"`
	// TLSKey specifies the TLS key for the domain used for installation.
	TLSKey string `json:"tlsKey"`
	// IsLocal indicates if the installation is on a local cluster.
	IsLocal      bool          `json:"isLocal"`
	LocalCluster *LocalCluster `json:"localCluster"`

	// Timeout specifies the time-out after which watching the installation progress stops.
	Timeout time.Duration `json:"timeout"`
	// NoWait determines if the Kyma installation should be waited to complete.
	NoWait bool `json:"noWait"`
	// Verbose enables displaying details of actions triggered.
	Verbose bool `json:"verbose"`
	// KubeconfigPath specifies the path to the kubeconfig file.
	KubeconfigPath string `json:"kubeconfigPath"`
}

// LocalCluster includes the configuration options of a local cluster.
type LocalCluster struct {
	// Provider specifies the provider of the local cluster.
	Provider string `json:"localProvider"`
	// Profile specifies the profile of the local cluster.
	Profile string `json:"localProfile"`
	// IP holds the IP of the local cluster.
	IP string `json:"localIP"`
	// VMDriver indicates the VM driver of the local cluster.
	VMDriver string `json:"localVMDriver"`
}
