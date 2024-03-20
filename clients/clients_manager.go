package clients

import (
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"net"
	"sync"
)

var manager *defaultClientsManager
var managerOnce sync.Once

func New(opts ...*ClientManagerOptions) ClientManagerInterface {
	managerOnce.Do(func() {
		opt := newOptions()
		if opts != nil && len(opts) == 1 && opts[0] != nil {
			opt = opt.merge(opts[0])
		}
		manager = &defaultClientsManager{
			mu:          sync.RWMutex{},
			clientMap:   map[string]ClientInterface{},
			tcpListener: nil,
			udpListener: nil,
			webListener: nil,
			opt:         opt,
		}
	})
	return manager
}

type defaultClientsManager struct {
	mu          sync.RWMutex
	clientMap   map[string]ClientInterface
	tcpListener *net.TCPListener
	udpListener net.Listener
	webListener *net.TCPListener
	opt         *ClientManagerOptions
}

func (m *defaultClientsManager) Len() int {
	//TODO implement me
	panic("implement me")
}

func (m *defaultClientsManager) List(start, end int) (int, clients_dto.ConnectionDatabase) {
	//TODO implement me
	panic("implement me")
}

func (m *defaultClientsManager) Start() error {
	//TODO implement me
	panic("implement me")
}

func (m *defaultClientsManager) Stop() {
	//TODO implement me
	panic("implement me")
}

func (m *defaultClientsManager) SetOptions(options *ClientManagerOptions) {
	//TODO implement me
	panic("implement me")
}
