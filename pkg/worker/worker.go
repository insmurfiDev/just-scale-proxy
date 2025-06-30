package worker

import (
	"context"

	"github.com/insmurfiDev/just-scale-proxy/internal/worker"
	"github.com/insmurfiDev/just-scale-proxy/pkg/common"
)

type Worker interface {
	Run(context.Context)
}

func NewWorker(cfg common.WorkerConfig) Worker {
	return worker.NewWorker(cfg)
}
