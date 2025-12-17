package docker

import (
	"github.com/docker/docker/client"
)

type Client struct {
	client.APIClient
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{APIClient: cli}, nil
}

func NewTestClient(mock client.APIClient) *Client {
	return &Client{APIClient: mock}
}
