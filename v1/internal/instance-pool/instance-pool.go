package instancepool

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/insmurfiDev/just-scale-proxy/v1/pkg/common"
)

// Хранилище подключенных микросервисов
type InstancePool struct {
	listener net.Listener

	conns map[net.Conn]common.Host
	mu    sync.RWMutex

	connectChan    chan common.Host
	disconenctChan chan common.Host
	msgChan        chan common.WorkerMsg

	cfg InstancePoolConfig
}

// Метод для обработки входящих соединений
func (p *InstancePool) Listen(ctx context.Context) {
	defer func() {
		p.listener.Close()
		close(p.connectChan)
		close(p.disconenctChan)
		close(p.msgChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := p.listener.Accept()

			if err != nil {
				slog.Warn("err handle connection", slog.Any("Error", err))
				continue
			}

			go p.handleConn(ctx, conn)
		}
	}

}

func (p *InstancePool) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	var host common.Host
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := conn.Read(buf)
			if err != nil {
				if p.cfg.OnDisconnected != nil {
					host, ok := p.getHostByConn(conn)
					if !ok {
						return
					}
					p.cfg.OnDisconnected(ctx, host, err)
				}
				p.disconnectConn(conn)
				return
			}

			message := string(buf[:n])

			if message == "ping" {
				conn.Write([]byte("pong"))
			} else {
				if host.IP != "" {
					select {
					case p.msgChan <- common.WorkerMsg{
						Msg:  buf[:n],
						Host: host,
					}:
					default:
					}
				} else {
					arr := strings.Split(message, ":")

					if len(arr) != 2 {
						continue
					}

					port, err := strconv.Atoi(arr[1])

					if err != nil {
						continue
					}
					host = common.Host{
						IP:   arr[0],
						Port: port,
					}
					p.saveConn(conn, host)

					if p.cfg.OnConnected != nil {
						p.cfg.OnConnected(ctx, host)
					}
				}
			}
		}
	}
}

func (p *InstancePool) getHostByConn(conn net.Conn) (common.Host, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.conns[conn]

	return v, ok
}

func (p *InstancePool) saveConn(conn net.Conn, host common.Host) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conns[conn] = host
	p.connectChan <- host
}

func (p *InstancePool) disconnectConn(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	host, ok := p.conns[conn]

	if ok {
		delete(p.conns, conn)
		p.disconenctChan <- host
	}
}

// Канал событий коннекта
func (p *InstancePool) ConnectChan() <-chan common.Host {
	return p.connectChan
}

// Канал событий дисконнекта
func (p *InstancePool) DisconnectChan() <-chan common.Host {
	return p.disconenctChan
}

func (p *InstancePool) Msg() <-chan common.WorkerMsg {
	return p.msgChan
}

type InstancePoolConfig struct {
	PossibleWorkersCount int
	IP                   string
	Port                 int
	OnConnected          func(context.Context, common.Host)
	OnDisconnected       func(context.Context, common.Host, error)
}

func NewInstancePool(cfg InstancePoolConfig) *InstancePool {

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.IP, cfg.Port))
	if err != nil {
		panic("can not listen: " + err.Error())
	}

	return &InstancePool{
		listener:       ln,
		conns:          make(map[net.Conn]common.Host, cfg.PossibleWorkersCount),
		connectChan:    make(chan common.Host),
		disconenctChan: make(chan common.Host),
		msgChan:        make(chan common.WorkerMsg),
		cfg:            cfg,
	}
}
