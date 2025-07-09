package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/insmurfiDev/just-scale-proxy/tests/generated/my_package"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/worker"
	"google.golang.org/grpc"
)

func main() {

	for i := range 1 {
		go conenctClient(9300 + i)
	}

	select {}
}

func conenctClient(port int) {
	ctx := context.Background()
	cfg := common.WorkerConfig{
		ProxyIP:       "0.0.0.0",
		ProxyPort:     9500,
		IP:            "0.0.0.0",
		Port:          port,
		RetryInterval: time.Second,
		Retry:         true,
		OnConnected: func(ctx context.Context) {
			fmt.Println("connected", port)
		},
		OnDisconnected: func(ctx context.Context) {
			fmt.Println("disconnected", port)
		},
	}

	w := worker.NewWorker(cfg)

	go w.Run(ctx)

	server, err := newServer(port)
	if err != nil {
		panic(err)
	}

	go server.Run()
}

type Server struct {
	my_package.UnimplementedMyServiceServer
	listener net.Listener
	server   *grpc.Server
	port     int
}

func (s *Server) Run() {
	my_package.RegisterMyServiceServer(s.server, s)
	s.server.Serve(s.listener)
}

func (s *Server) SomeMethod(ctx context.Context, req *my_package.SomeReq) (*my_package.SomeRes, error) {
	slog.Info("handle method", slog.Int("port", s.port))
	return &my_package.SomeRes{
		Res: 1,
	}, nil
}

func newServer(port int) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()

	return &Server{
		listener: l,
		server:   s,
		port:     port,
	}, nil
}
