package service

import (
	"context"

	pb "github.com/warpgr/test_task/proto"
)

type (
	RateService interface {
		pb.RateServiceServer
	}

	Daemon interface {
		Run(ctx context.Context) <-chan error
	}
)
