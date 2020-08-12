module github.com/kyma-project/cli

go 1.12

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.6.0
	github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
	github.com/codegangsta/cli => github.com/urfave/cli v1.22.4
	// etcd and ugorji need to be versioned together and we need ot force the version from terraform 0.12.13 otherwise we have an ambiguous import
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.15+incompatible
	github.com/ugorji/go v0.0.0-20180813092308-00b869d2f4a5 => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	// Docker client has an issue on windows with the latest sys package, we have to fix the version
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190710143415-6ec70d6a5542
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.15-0.20200113171025-3fe6c5262873 // indirect
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/briandowns/spinner v1.7.0
	github.com/containerd/containerd v1.3.6 // indirect
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/daviddengcn/go-colortext v0.0.0-20180409174941-186a3d44e920
	github.com/docker/cli v0.0.0-20190925022749-754388324470
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20191101170500-ac7306503d23
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/fatih/color v1.7.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/golangplus/testing v0.0.0-20180327235837-af21d9c3145e // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kyma-incubator/hydroform/install v0.0.0-20200812115205-1299dd4d0c6c
	github.com/kyma-incubator/hydroform/provision v0.0.0-20200803123159-99d6ef03bf0c
	github.com/kyma-incubator/octopus v0.0.0-20191009105757-2e9d86cd9967
	github.com/kyma-project/kyma v0.5.1-0.20200211132707-0a36a0f31d7e
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.0-rc91 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.4 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/sys v0.0.0-20200722175500-76b94024e4b6 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200722002428-88e341933a54 // indirect
	google.golang.org/grpc v1.29.1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20191114100237-2cd11237263f
	k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/yaml v1.1.0
)
