package modulesv2

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/out"
)

type PullService struct{}

func NewPullService() *PullService {
	return &PullService{}
}

func (s *PullService) Run(ctx context.Context) error {
	out.Msgln("Hello, there! I'm just a dummy PullService! Can you imagine haha")

	return nil
}
