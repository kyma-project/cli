module github.com/kyma-project/cli

go 1.14

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.6.0
	github.com/kyma-incubator/hydroform/function => github.com/pPrecel/hydroform/function v0.0.0-20200928191448-557119d735e9
	// this is needed for terraform to work with the k8s 0.18 APIs, we should be able to remove it once we have terraform 0.13+
	github.com/terraform-providers/terraform-provider-openstack => github.com/terraform-providers/terraform-provider-openstack v1.20.0
	// grpc need to be compatible with direct dependencies in terraform (>=v1.29.1)
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.15-0.20200113171025-3fe6c5262873 // indirect
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/containerd/containerd v1.3.6 // indirect
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/daviddengcn/go-colortext v1.0.0
	github.com/docker/cli v0.0.0-20200130152716-5d0cf8839492
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/fatih/color v1.9.0
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/kyma-incubator/hydroform/function v0.0.0-00010101000000-000000000000
	github.com/kyma-incubator/hydroform/install v0.0.0-20200922142757-cae045912c90
	github.com/kyma-incubator/hydroform/provision v0.0.0-20200922142757-cae045912c90
	github.com/kyma-incubator/octopus v0.0.0-20200922132758-2b721e93b58b
	github.com/kyma-project/kyma/components/kyma-operator v0.0.0-20200923072402-e911fd636ae6
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.0.0-rc91 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	gotest.tools v2.2.0+incompatible
	istio.io/api v0.0.0-20200911191701-0dc35ad5c478
	istio.io/client-go v0.0.0-20200807182027-d287a5abb594
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
	sigs.k8s.io/yaml v1.2.0
)
