package loadbalancer

import (
	"github.com/insmurfiDev/just-scale-proxy/internal/loadbalancer"
)

type LoadBalancer interface {
	Next() int
	SetSize(int)
	GetSize() int
}

func NewRoundRobbin() LoadBalancer {
	return &loadbalancer.RoundRobbin{}
}
