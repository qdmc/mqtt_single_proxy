// Package clients  mqtt客户端及客户端管理器
package clients

import (
	"github.com/qdmc/mqtt_packet"
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"github.com/qdmc/mqtt_single_proxy/enmu"
)

// ConnectedCallback      链接成功的回调
type ConnectedCallback func(string)

// PacketCallbackHandle   报文回调Handle
type PacketCallbackHandle func(string, mqtt_packet.ControlPacketInterface)

// DisConnectCallbackHandle  客户端断开回调Handle
type DisConnectCallbackHandle func(*clients_dto.ConnectionDatabase)

// HandshakeHandle          握手校验Handle
type HandshakeHandle func(clients_dto.ConnectionHandshakeDatabase) enmu.HandshakeResult

// ClientInterface   客户端通用接口
type ClientInterface interface {
	GetId() string                                  // 返回ClientId
	GetDataBase() clients_dto.ConnectionDatabase    // 返回当前状态
	AsyncDoConnection()                             // 异步处理Tcp长链接
	SetStatistics(bool)                             // 设置是否开启数据统计
	SetPacketHandle(PacketCallbackHandle)           // 配置报文回调
	SetDisConnectCallback(DisConnectCallbackHandle) // 配置断开回调
	DisConnect()                                    // 断开链接
	GetProtocol() enmu.ClientProtocol
}

type ClientManagerInterface interface {
	Len() int                                                  // 返回客户端总数
	List(start, end int) (int, clients_dto.ConnectionDatabase) // 返回客户端列表
	Start() error
	Stop()
	SetOptions(*ClientManagerOptions)
}
