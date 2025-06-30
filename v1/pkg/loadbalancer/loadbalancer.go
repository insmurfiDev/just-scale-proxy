package loadbalancer

import (
	"github.com/insmurfiDev/just-scale-proxy/v1/internal/loadbalancer"
)

type LoadBalancer interface {
	Next() int
	SetSize(int)
	GetSize() int
}

func NewRoundRobbin() LoadBalancer {
	return &loadbalancer.RoundRobbin{}
}
