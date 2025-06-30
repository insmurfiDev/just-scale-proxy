package worker

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/insmurfiDev/just-scale-proxy/pkg/common"
)

type Worker struct {
	host      string
	proxyHost string
	cfg       common.WorkerConfig

	conn net.Conn
}

func (w *Worker) Connect(ctx context.Context) error {
	if w.conn != nil {
		_ = w.conn.Close()
	}

	conn, err := net.Dial("tcp", w.proxyHost)
	if err != nil {
		return fmt.Errorf("failed to connect to proxy: %w", err)
	}
	w.conn = conn

	if _, err := conn.Write([]byte(w.host)); err != nil {
		return fmt.Errorf("failed to write host info: %w", err)
	}

	if w.cfg.OnConnected != nil {
		w.cfg.OnConnected(ctx)
	}

	return nil
}

func (w *Worker) Run(ctx context.Context) {

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if w.conn == nil {
					if w.cfg.OnDisconnected != nil {
						w.cfg.OnDisconnected(ctx)
					}
					if w.cfg.Retry {
						if err := w.Connect(ctx); err != nil {
							time.Sleep(w.cfg.RetryInterval)
							continue
						}
					}
				}

				reader := bufio.NewReader(w.conn)
				_, err := reader.ReadByte()

				if err != nil {
					w.conn.Close()
					w.conn = nil
					time.Sleep(w.cfg.RetryInterval)
					continue
				}
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
