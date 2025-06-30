package proxy

import (
	"context"

	p "github.com/insmurfiDev/just-scale-proxy/v1/internal/proxy"
	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
)

type Proxy[Client comparable] interface {
	GetClient(context.Context, any) Client
	GetAnyClient() Client
	Run(context.Context)
}

func NewProxy[Client comparable](cfg common.ProxyConfig[Client]) Proxy[Client] {
	return p.NewProxy[Client](cfg)
}
