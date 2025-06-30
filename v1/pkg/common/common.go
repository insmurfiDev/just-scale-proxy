package common

import (
	"context"
	"time"

	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/loadbalancer"
)

// Хост микросервиса
type Host struct {
	IP   string
	Port int
}

// Хеш функция для запроса
type HashFn func(ctx context.Context, req any) int

type CreateClientFn[Client any] func(ctx context.Context, ip string, port int) (Client, error)

type ProxyConfig[Client comparable] struct {
	HashFn              HashFn
	PossibleClientCount int
	CreateClientFn      CreateClientFn[Client]
	LoadBalancer        loadbalancer.LoadBalancer

	IP   string
	Port int

	OnConnected    func(context.Context, Host)
	OnDisconnected func(context.Context, Host)
	OnError        func(context.Context, Host, error)
}

type WorkerConfig struct {
	ProxyIP   string
	ProxyPort int

	IP   string
	Port int

	RetryInterval time.Duration
	Retry         bool

	OnConnected    func(context.Context)
	OnDisconnected func(context.Context)
}
