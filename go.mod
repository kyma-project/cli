module github.com/kyma-project/cli

go 1.14

replace (
	//TODO: remove this part as Helm 3.5.4 got released (see dep in Hydroform API)
	//see https://github.com/helm/helm/issues/9354 + https://github.com/helm/helm/pull/9492
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v20.10.5+incompatible

	github.com/hashicorp/consul v0.0.0-20171026175957-610f3c86a089 => github.com/hashicorp/consul/sdk v0.7.0
	github.com/hashicorp/consul/api v1.3.0 => github.com/hashicorp/consul/api v0.0.0-20191112221531-8742361660b6
	// this is needed for terraform to work with the k8s 0.18 APIs, we should be able to remove it once we have terraform 0.13+
	github.com/terraform-providers/terraform-provider-openstack => github.com/terraform-providers/terraform-provider-openstack v1.20.0
	// grpc need to be compatible with direct dependencies in terraform (>=v1.29.1)
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/briandowns/spinner v1.12.0
	github.com/containerd/cgroups v0.0.0-20201119153540-4cbc285b3327 // indirect
	github.com/daviddengcn/go-colortext v1.0.0
	github.com/docker/cli v20.10.3+incompatible
	github.com/docker/docker v20.10.3+incompatible
	github.com/fatih/color v1.10.0
	github.com/kyma-incubator/hydroform/function v0.0.0-20210319130226-926c73801013
	github.com/kyma-incubator/hydroform/install v0.0.0-20200922142757-cae045912c90
	github.com/kyma-incubator/hydroform/parallel-install v0.0.0-20210409075848-dbbbeec7084e
	github.com/kyma-incubator/hydroform/provision v0.0.0-20201124135641-ca1a1a00c935
	github.com/kyma-incubator/octopus v0.0.0-20200922132758-2b721e93b58b
	github.com/kyma-project/kyma/components/kyma-operator v0.0.0-20201125092745-687c943ac940
	github.com/magiconair/properties v1.8.1
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.22.5 // indirect
	go.uber.org/zap v1.16.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	istio.io/api v0.0.0-20200911191701-0dc35ad5c478
	istio.io/client-go v0.0.0-20200807182027-d287a5abb594
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/yaml v1.2.0
)
