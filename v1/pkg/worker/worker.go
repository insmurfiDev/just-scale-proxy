package worker

import (
	"context"

	"github.com/insmurfiDev/just-scale-proxy/v1/internal/worker"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
)

type Worker interface {
	Run(context.Context)
}

func NewWorker(cfg common.WorkerConfig) Worker {
	return worker.NewWorker(cfg)
}
