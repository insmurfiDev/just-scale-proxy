package worker

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
)

type Worker struct {
	host      string
	proxyHost string
	cfg       common.WorkerConfig

	isConnected atomic.Bool

	conn net.Conn
}

func (w *Worker) onError(ctx context.Context, err error) {
	if w.cfg.OnError != nil {
		w.cfg.OnError(ctx, err)
	}
}

func (w *Worker) onConnected(ctx context.Context) {
	if w.cfg.OnConnected != nil {
		w.cfg.OnConnected(ctx)
	}
}

func (w *Worker) onDisconnected(ctx context.Context) {
	if w.cfg.OnDisconnected != nil {
		w.cfg.OnDisconnected(ctx)
	}
}

func (w *Worker) Disconnect(ctx context.Context) {
	if w.isConnected.Load() {
		w.onDisconnected(ctx)
	}

	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}
	w.isConnected.Store(false)
}

func (w *Worker) Connect(ctx context.Context) {
	conn, err := net.Dial("tcp", w.proxyHost)
	if err != nil {
		w.onError(ctx, fmt.Errorf("failed to connect to proxy: %w", err))
		w.Disconnect(ctx)
		return
	}
	w.conn = conn

	if _, err := conn.Write([]byte(w.host)); err != nil {
		w.onError(ctx, fmt.Errorf("failed to write host info: %w", err))
		w.Disconnect(ctx)
		return
	}

	w.isConnected.Store(true)
	w.onConnected(ctx)
}

func (w *Worker) handleRetry(ctx context.Context) {
	if w.cfg.Retry {
		for !w.isConnected.Load() {
			time.Sleep(w.cfg.RetryInterval)
			w.Connect(ctx)
		}
	}
}

func (w *Worker) Run(ctx context.Context) {

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		w.Connect(ctx)
		if !w.isConnected.Load() {
			w.handleRetry(ctx)
		}

		buf := make([]byte, 1024)
		for {
			if w.conn == nil {
				continue
			}
			_, err := w.conn.Read(buf)
			if err != nil {
				w.onError(ctx, fmt.Errorf("failed to read from proxy: %w", err))
				w.Disconnect(ctx)
				w.handleRetry(ctx)
				continue
			}
		}

	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if w.conn != nil {
					w.conn.Write([]byte("ping"))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

func NewWorker(cfg common.WorkerConfig) *Worker {
	return &Worker{
		host:      fmt.Sprintf("%s:%d", cfg.IP, cfg.Port),
		proxyHost: fmt.Sprintf("%s:%d", cfg.ProxyIP, cfg.ProxyPort),
		cfg:       cfg,
	}
}
