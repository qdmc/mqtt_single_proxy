package clients

import (
	"github.com/qdmc/mqtt_packet"
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"github.com/qdmc/mqtt_single_proxy/enmu"
	"net"
	"sort"
	"sync"
)

var manager *defaultClientManager
var managerOnce sync.Once

func NewClientManager(opts ...*ClientManagerOptions) ClientManagerInterface {
	managerOnce.Do(func() {
		opt := newOptions()
		if opts != nil && len(opts) == 1 && opts[0] != nil {
			opt = opt.merge(opts[0])
		}
		manager = &defaultClientManager{
			mu:          sync.RWMutex{},
			clientMap:   map[string]clientInterface{},
			tcpListener: nil,
			udpListener: nil,
			webListener: nil,
			opt:         opt,
		}
	})
	return manager
}

type defaultClientManager struct {
	mu          sync.RWMutex
	clientMap   map[string]clientInterface
	tcpListener *net.TCPListener
	udpListener net.Listener
	webListener *net.TCPListener
	opt         *ClientManagerOptions
	isStart     bool
}

func (m *defaultClientManager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clientMap)
}

func (m *defaultClientManager) List(start, end int) (int, []clients_dto.ConnectionDatabase) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ds []clients_dto.ConnectionDatabase
	var cs clients
	length := len(m.clientMap)
	if length == 0 {
		return length, ds
	}
	for _, client := range m.clientMap {
		cs = append(cs, client)
	}
	sort.Sort(cs)
	for index, c := range cs {
		if index >= start && index < end {
			ds = append(ds, c.GetDataBase())
		}
	}
	return length, ds
}

func (m *defaultClientManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var err error
	defer func() {
		if err != nil {
			m.isStart = true
		} else {

		}
	}()
	// todo
	return nil
}

func (m *defaultClientManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.isStart {
		return nil
	}
	var err error
	err = m.tcpListener.Close()
	if m.udpListener != nil {
		err = m.udpListener.Close()
	}
	if m.webListener != nil {
		err = m.webListener.Close()
	}
	for _, client := range m.clientMap {
		go client.DisConnect(true)
	}
	m.clientMap = map[string]clientInterface{}
	m.isStart = false
	return err
}

func (m *defaultClientManager) SetOptions(options *ClientManagerOptions) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isStart {
		return
	}
	m.opt = m.opt.merge(options)
}
func (m *defaultClientManager) CloseOnce(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, ok := m.clientMap[id]; ok {
		client.DisConnect()
		return nil
	}
	return enmu.NotFoundClientError
}
func (m *defaultClientManager) GetOnce(id string) (*clients_dto.ConnectionDatabase, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, ok := m.clientMap[id]; ok {
		cb := client.GetDataBase()
		return &cb, nil
	}
	return nil, enmu.NotFoundClientError
}
func (m *defaultClientManager) SendPacketOnce(id string, p mqtt_packet.ControlPacketInterface) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if client, ok := m.clientMap[id]; ok {
		return client.WritePacketOnce(p)
	}
	return 0, enmu.NotFoundClientError
}
func (m *defaultClientManager) doDisConnectCb(cd *clients_dto.ConnectionDatabase) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cd == nil {
		return
	}
	if _, ok := m.clientMap[cd.Id]; ok {
		delete(m.clientMap, cd.Id)
		if m.opt.DisConnectCb != nil {
			go m.opt.DisConnectCb(cd)
		}
	}
}
func (m *defaultClientManager) doPacketCb(id string, p mqtt_packet.ControlPacketInterface) {
	if m.opt.PacketCb != nil {
		go m.opt.PacketCb(id, p)
	}
}
func (m *defaultClientManager) doConnectedCb(id string) {
	if m.opt.ConnectedCb != nil {
		go m.opt.ConnectedCb(id)
	}
}

type clients []clientInterface

func (cs clients) Len() int {
	return len(cs)
}
func (cs clients) Less(i, j int) bool {
	return cs[i].GetId() > cs[j].GetId()
}

func (cs clients) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}
