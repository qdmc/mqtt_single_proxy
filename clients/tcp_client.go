package clients

import (
	"github.com/qdmc/mqtt_packet"
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"github.com/qdmc/mqtt_single_proxy/enmu"
	"net"
	"sync/atomic"
	"time"
)

type tcpClient struct {
	id            string
	status        bool
	disConnectCb  DisConnectCallbackHandle
	packetCb      PacketCallbackHandle
	connectedNano int64   // 链接开始时间
	closeNano     int64   // 链接断开时间
	writeLength   *uint64 // 发送的数据长度
	readLength    *uint64 // 接收的数据长度
	isStatistics  bool    // 是否开启流量统计,默认为false
	e             error
	isNoCb        bool
	conn          net.Conn
	stopChan      chan struct{}
	tr            *time.Timer
	t             time.Duration
}

func newTcpClient(id string, conn net.Conn, statistics ...bool) *tcpClient {
	isStatistics := false
	if statistics != nil && len(statistics) == 1 && statistics[0] {
		isStatistics = true
	}
	var rl, wl uint64
	return &tcpClient{
		id:            id,
		status:        false,
		disConnectCb:  nil,
		packetCb:      nil,
		connectedNano: time.Now().UnixNano(),
		closeNano:     0,
		writeLength:   &wl,
		readLength:    &rl,
		isStatistics:  isStatistics,
		conn:          conn,
		stopChan:      make(chan struct{}, 1),
		t:             time.Duration(60) * time.Second,
	}
}

func (c *tcpClient) GetId() string {
	return c.id
}

func (c *tcpClient) GetDataBase() clients_dto.ConnectionDatabase {
	return clients_dto.ConnectionDatabase{
		Id:            c.id,
		Protocol:      c.GetProtocol(),
		ConnectedNano: c.connectedNano,
		CloseNano:     c.closeNano,
		WriteLength:   atomic.LoadUint64(c.writeLength),
		ReadLength:    atomic.LoadUint64(c.readLength),
		Status:        c.status,
		IsStatistics:  c.isStatistics,
		Err:           c.e,
	}
}

func (c *tcpClient) AsyncDoConnection() {
	c.status = true
	var err error
	defer func() {
		c.status = false
		c.doDisconnect(err)
	}()
	c.tr = time.NewTimer(c.t)
	for {
		select {
		case <-c.stopChan:
			err = nil
			return
		case <-c.tr.C:
			err = enmu.ClientHeartTimeoutError
			return
		default:
			readLen, p, readErr := mqtt_packet.ReadOnce(c.conn)
			if readErr != nil {
				err = enmu.ClientReadConnectionError
				return
			}
			c.tr.Reset(c.t)
			if c.isStatistics {
				atomic.AddUint64(c.readLength, uint64(readLen))
			}
			go c.doPacket(p)
			continue
		}
	}
}
func (c *tcpClient) SetTimeOut(t int64) {
	if !c.status {
		if t >= 10 && t <= 300 {
			c.t = time.Duration(t) * time.Second
		}
	}
}
func (c *tcpClient) SetStatistics(b bool) {
	if !c.status {
		c.isStatistics = b
	}
}

func (c *tcpClient) SetPacketHandle(handle PacketCallbackHandle) {
	if !c.status {
		c.packetCb = handle
	}
}

func (c *tcpClient) SetDisConnectCallback(handle DisConnectCallbackHandle) {
	if !c.status {
		c.disConnectCb = handle
	}
}

func (c *tcpClient) DisConnect(isNoCb ...bool) {
	if isNoCb != nil && len(isNoCb) == 1 && isNoCb[0] {
		c.isNoCb = true
	}
	close(c.stopChan)
}

func (c *tcpClient) GetProtocol() enmu.ClientProtocol {
	return enmu.TcpProtocol
}
func (c *tcpClient) WritePacketOnce(p mqtt_packet.ControlPacketInterface) (int64, error) {
	if p == nil {
		return 0, enmu.PacketEmptyError
	}
	if c.status {
		length, err := p.Write(c.conn)
		if err != nil {
			return 0, err
		}
		if c.isStatistics {
			atomic.AddUint64(c.writeLength, uint64(length))
		}
		return length, nil
	} else {
		return 0, enmu.ClientDisconnectError
	}
}
func (c *tcpClient) doPacket(p mqtt_packet.ControlPacketInterface) {
	if p == nil || c.packetCb == nil {
		return
	}
	go c.packetCb(c.id, p)
}

func (c *tcpClient) doDisconnect(err error) {
	if err != nil {
		c.e = err
	}
	if c.tr != nil {
		c.tr.Stop()
	}
	if !c.isNoCb && c.disConnectCb != nil {
		db := c.GetDataBase()
		go c.disConnectCb(&db)
	}
}
