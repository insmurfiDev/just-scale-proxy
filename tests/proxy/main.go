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

type Client struct {
	Client my_package.MyServiceClient
	Port   int64
}

//go:generate protoc --go_out=../generated --go-grpc_out=../generated --proto_path=../ test.proto
func main() {

	ctx := context.Background()

	cfg := common.ProxyConfig[*Client]{
		HashFn: func(ctx context.Context, req any) int {
			id, ok := req.(int64)
			if !ok {
				fmt.Println("err parse to int", req)
				return 0
			}

			if id < 1 {
				return 0
			}

			if id < 2 {
				return 1
			}

			if id < 3 {
				return 2
			}

			if id < 4 {
				return 3
			}

			if id < 5 {
				return 4
			}

			return int(id)
		},
		CreateClientFn: func(ctx context.Context, ip string, port int) (*Client, error) {
			cc, err := grpc.NewClient(
				fmt.Sprintf("%s:%d", ip, port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
			)

			if err != nil {
				return nil, err
			}

			client := my_package.NewMyServiceClient(cc)
			return &Client{
				Client: client,
				Port:   int64(port),
			}, nil

		},
		LoadBalancer: loadbalancer.NewRoundRobbin(),
		OnConnected: func(ctx context.Context, h common.Host) {
			fmt.Println("connected", h.IP, h.Port)
		},
		OnDisconnected: func(ctx context.Context, h common.Host, err error) {
			fmt.Println("disconnected", h.IP, h.Port)
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
	proxy    proxy.Proxy[*Client]
}

func (s *Server) Run() {
	my_package.RegisterMyServiceServer(s.server, s)
	s.server.Serve(s.listener)
}

func (s *Server) SomeMethod(ctx context.Context, req *my_package.SomeReq) (*my_package.SomeRes, error) {
	// client := s.proxy.GetClient(ctx, req.Id) // sticky
	client := s.proxy.GetAnyClient() // round
	port := client.Port
	slog.Info("Перенаправил запрос", slog.Int64("id", req.Id), slog.Int64("port", port))
	return client.Client.SomeMethod(ctx, req)
}

func newServer(proxy proxy.Proxy[*Client]) (*Server, error) {
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
