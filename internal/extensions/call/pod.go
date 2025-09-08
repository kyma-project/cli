package call

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/portforward"
)

type PodCaller struct {
	ctx          context.Context
	client       kube.Client
	podSelector  map[string]string
	podNamespace string
	podPort      string
}

func NewPodCaller(ctx context.Context, client kube.Client, podNamespace string, podSelector map[string]string, podPort string) *PodCaller {
	return &PodCaller{
		ctx:          ctx,
		client:       client,
		podSelector:  podSelector,
		podNamespace: podNamespace,
		podPort:      podPort,
	}
}

func (c *PodCaller) Call(method, path string, parameters map[string]string) ([]byte, clierror.Error) {
	targetPod, err := resources.GetPodForSelector(
		c.ctx,
		c.client.Static(),
		c.podNamespace,
		c.podSelector,
	)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to get target pod"))
	}

	req, err := buildRequest(method, targetPod.GetName(), targetPod.GetNamespace(), c.podPort, path, parameters)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to build request"))
	}

	resp, err := portforward.DoRequest(
		c.client.RestConfig(),
		targetPod.GetName(),
		targetPod.GetNamespace(),
		c.podPort,
		req,
	)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to send request to target pod"))
	}
	defer resp.Body.Close()

	return decodeResponse(resp)
}

func buildRequest(method, podName, podNamespace, podPort, path string, parameters map[string]string) (*http.Request, error) {
	address := fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", podName, podNamespace, podPort)
	req, err := http.NewRequest(method, address, nil)
	if err != nil {
		return nil, err
	}

	req.URL.Path = path

	query := req.URL.Query()
	for k, v := range parameters {
		query.Add(k, v)
	}

	req.URL.RawQuery = query.Encode()
	return req, nil
}
