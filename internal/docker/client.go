package docker

import (
	"github.com/docker/docker/client"
)

type Client struct {
	client client.APIClient
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{client: cli}, nil
}

func NewTestClient(mock client.APIClient) *Client {
	return &Client{client: mock}
}
