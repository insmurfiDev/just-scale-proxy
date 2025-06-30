package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/insmurfiDev/just-scale-proxy/tests/generated/my_package"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/loadbalancer"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate protoc --go_out=../generated --go-grpc_out=../generated --proto_path=../ test.proto
func main() {

	ctx := context.Background()

	cfg := common.ProxyConfig[my_package.MyServiceClient]{
		HashFn: func(ctx context.Context, req any) int {
			id, ok := req.(int64)
			if !ok {
				fmt.Println("err parse to int", req)
				return 0
			}
			return int(id)
		},
		CreateClientFn: func(ctx context.Context, ip string, port int) (my_package.MyServiceClient, error) {
			cc, err := grpc.NewClient(
				fmt.Sprintf("%s:%d", ip, port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
			)

			if err != nil {
				return nil, err
			}

			client := my_package.NewMyServiceClient(cc)
			return client, nil

		},
		LoadBalancer: loadbalancer.NewRoundRobbin(),
		OnConnected: func(ctx context.Context, h common.Host) {
			fmt.Println("connected", h.IP, h.Port)
		},
		IP:   "0.0.0.0",
		Port: 9500,
	}

	p := proxy.NewProxy(cfg)

	go p.Run(ctx)

	server, err := newServer(p)
	if err != nil {
		panic(err)
	}
	go server.Run()

	select {
	case <-ctx.Done():
		return
	}

}

type Server struct {
	my_package.UnimplementedMyServiceServer
	listener net.Listener
	server   *grpc.Server
	proxy    proxy.Proxy[my_package.MyServiceClient]
}

func (s *Server) Run() {
	my_package.RegisterMyServiceServer(s.server, s)
	s.server.Serve(s.listener)
}

func (s *Server) SomeMethod(ctx context.Context, req *my_package.SomeReq) (*my_package.SomeRes, error) {
	slog.Info("get client")
	client := s.proxy.GetClient(ctx, req.Id)
	slog.Info("got client", client)
	return client.SomeMethod(ctx, req)
}

func newServer(proxy proxy.Proxy[my_package.MyServiceClient]) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 9505))

	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()

	return &Server{
		listener: l,
		server:   s,
		proxy:    proxy,
	}, nil
}
