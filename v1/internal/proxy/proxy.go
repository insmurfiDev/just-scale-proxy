package internal

import (
	"context"

	clientpool "github.com/insmurfiDev/just-scale-proxy/v1/internal/client-pool"
	instancepool "github.com/insmurfiDev/just-scale-proxy/v1/internal/instance-pool"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/loadbalancer"
)

type Proxy[Client comparable] struct {
	clientPool   *clientpool.ClientPool[Client]
	instancePool *instancepool.InstancePool

	createClientFn common.CreateClientFn[Client]

	loadBalancer loadbalancer.LoadBalancer
}

func (p *Proxy[Client]) GetClient(ctx context.Context, req any) Client {
	return p.clientPool.Get(ctx, req).Client
}

func (p *Proxy[Client]) GetAnyClient() Client {
	idx := p.loadBalancer.Next()

	return p.clientPool.GetByIdx(idx).Client
}

func (p *Proxy[Client]) Run(ctx context.Context) {
	go p.instancePool.Listen(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case host, ok := <-p.instancePool.ConnectChan():
				if !ok {
					return
				}
				client, err := p.createClientFn(ctx, host.IP, host.Port)
				if err != nil {
					continue
				}

				p.clientPool.Add(clientpool.ClientInfo[Client]{
					Client: client,
					Host:   host,
				})
				newSize := p.loadBalancer.GetSize() + 1
				p.loadBalancer.SetSize(newSize)
			case host, ok := <-p.instancePool.DisconnectChan():
				if !ok {
					return
				}
				p.clientPool.DeleteByHost(ctx, host)
				newSize := p.loadBalancer.GetSize() - 1
				p.loadBalancer.SetSize(newSize)
			}

		}
	}()

}

func (p *Proxy[Client]) GetClientByHost(host common.Host) (Client, bool) {
	info, ok := p.clientPool.GetByHost(host)

	return info.Client, ok
}

func NewProxy[Client comparable](cfg common.ProxyConfig[Client]) *Proxy[Client] {
	return &Proxy[Client]{
		clientPool: clientpool.NewClientPool[Client](cfg.PossibleClientCount, cfg.HashFn),
		instancePool: instancepool.NewInstancePool(instancepool.InstancePoolConfig{
			PossibleWorkersCount: cfg.PossibleClientCount,
			IP:                   cfg.IP,
			Port:                 cfg.Port,
			OnConnected:          cfg.OnConnected,
			OnDisconnected:       cfg.OnDisconnected,
		}),
		createClientFn: cfg.CreateClientFn,
		loadBalancer:   cfg.LoadBalancer,
	}
}
