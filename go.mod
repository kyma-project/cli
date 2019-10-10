module github.com/kyma-project/cli

go 1.12

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

require (
	cloud.google.com/go v0.46.3 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/briandowns/spinner v1.7.0
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6 // indirect
	github.com/daviddengcn/go-colortext v0.0.0-20180409174941-186a3d44e920
	github.com/fatih/color v1.7.0
	github.com/fsouza/go-dockerclient v1.4.5-0.20190927145838-348231a36c19
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/golangplus/testing v0.0.0-20180327235837-af21d9c3145e // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kyma-incubator/octopus v0.0.0-20191009105757-2e9d86cd9967
	github.com/kyma-project/kyma v0.5.1-0.20190909070658-69599d4a33a2
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/olekukonko/tablewriter v0.0.1
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/net v0.0.0-20191003171128-d98b1b443823 // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	google.golang.org/appengine v1.6.4 // indirect
	google.golang.org/genproto v0.0.0-20191002211648-c459b9ce5143 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d // indirect
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
)
