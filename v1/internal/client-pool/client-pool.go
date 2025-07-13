package clientpool

import (
	"context"
	"sync"

	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
)

// Информация о подключенном клиенте
type ClientInfo[Client comparable] struct {
	// Клиент
	Client Client

	// Хост клиента
	Host common.Host
}

// Хланилище подключенных клиентов
type ClientPool[Client comparable] struct {
	clients []ClientInfo[Client]

	hashFn common.HashFn

	mu sync.RWMutex
}

func NewClientPool[Client comparable](
	possibleClientsCount int,
	hashFn common.HashFn,
) *ClientPool[Client] {
	return &ClientPool[Client]{
		clients: make([]ClientInfo[Client], 0, possibleClientsCount),
		hashFn:  hashFn,
	}
}

// Метод для добавления клиента в хранилище
func (c *ClientPool[Client]) Add(client ClientInfo[Client]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clients = append(c.clients, client)
}

// Метод для получения клиента из хранилища
//
// Получает клиента используя хеш-функцию
func (c *ClientPool[Client]) Get(ctx context.Context, req any) ClientInfo[Client] {
	hashFn := c.hashFn
	hash := hashFn(ctx, req)

	c.mu.RLock()
	idx := hash % len(c.clients)
	c.mu.RUnlock()
	return c.GetByIdx(idx)
}

// Метод для удаления клиента по хосту
func (c *ClientPool[Client]) DeleteByHost(ctx context.Context, host common.Host) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < len(c.clients); i++ {
		client := c.clients[i]
		if client.Host == host {
			c.clients = append(c.clients[:i], c.clients[i+1:]...)
			return
		}

	}

}

// Получает клиента по индексу
func (c *ClientPool[Client]) GetByIdx(idx int) ClientInfo[Client] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.clients[idx]
}
