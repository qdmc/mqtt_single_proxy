package clients

import (
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"github.com/qdmc/mqtt_single_proxy/enmu"
)

type tcpClient struct {
	status        bool
	disConnectCb  DisConnectCallbackHandle
	packetCb      PacketCallbackHandle
	connectedNano int64  // 链接开始时间
	closeNano     int64  // 链接断开时间
	writeLength   uint64 // 发送的数据长度
	readLength    uint64 // 接收的数据长度
	isStatistics  bool   // 是否开启流量统计,默认为false
}

func (c *tcpClient) GetId() string {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) GetDataBase() clients_dto.ConnectionDatabase {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) AsyncDoConnection() {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) SetStatistics(b bool) {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) SetPacketHandle(handle PacketCallbackHandle) {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) SetDisConnectCallback(handle DisConnectCallbackHandle) {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) DisConnect(isNoCb ...bool) {
	//TODO implement me
	panic("implement me")
}

func (c *tcpClient) GetProtocol() enmu.ClientProtocol {
	return enmu.TcpProtocol
}
