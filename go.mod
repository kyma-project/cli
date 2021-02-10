module github.com/kyma-project/cli

go 1.14

replace (
	github.com/hashicorp/consul v0.0.0-20171026175957-610f3c86a089 => github.com/hashicorp/consul/sdk v0.7.0
	github.com/hashicorp/consul/api v1.3.0 => github.com/hashicorp/consul/api v0.0.0-20191112221531-8742361660b6
	// this is needed for terraform to work with the k8s 0.18 APIs, we should be able to remove it once we have terraform 0.13+
	github.com/terraform-providers/terraform-provider-openstack => github.com/terraform-providers/terraform-provider-openstack v1.20.0
	// grpc need to be compatible with direct dependencies in terraform (>=v1.29.1)
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/client-go => k8s.io/client-go v0.18.9
)

require (
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/Microsoft/hcsshim v0.8.10 // indirect
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/briandowns/spinner v1.12.0
	github.com/containerd/cgroups v0.0.0-20201119153540-4cbc285b3327 // indirect
	github.com/containerd/containerd v1.4.2 // indirect
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/daviddengcn/go-colortext v1.0.0
	github.com/docker/cli v0.0.0-20200130152716-5d0cf8839492
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/fatih/color v1.10.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/kyma-incubator/hydroform/function v0.0.0-20201229063132-d2abcd251425
	github.com/kyma-incubator/hydroform/install v0.0.0-20200922142757-cae045912c90
	github.com/kyma-incubator/hydroform/parallel-install v0.0.0-20210204141431-7d490c516314
	github.com/kyma-incubator/hydroform/provision v0.0.0-20201124135641-ca1a1a00c935
	github.com/kyma-incubator/octopus v0.0.0-20200922132758-2b721e93b58b
	github.com/kyma-project/kyma/components/kyma-operator v0.0.0-20201125092745-687c943ac940
	github.com/magiconair/properties v1.8.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/runc v1.0.0-rc91 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	go.opencensus.io v0.22.5 // indirect
	golang.org/x/sys v0.0.0-20201126233918-771906719818 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.3 // indirect
	istio.io/api v0.0.0-20200911191701-0dc35ad5c478
	istio.io/client-go v0.0.0-20200807182027-d287a5abb594
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v8.0.0+incompatible
	sigs.k8s.io/yaml v1.2.0
)
