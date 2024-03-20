package clients_dto

import (
	"github.com/qdmc/mqtt_single_proxy/enmu"
	"net"
)

// ConnectionDatabase  链接数据统计
type ConnectionDatabase struct {
	Id            string // sessionId
	Protocol      enmu.ClientProtocol
	ConnectedNano int64  // 链接开始时间
	CloseNano     int64  // 链接断开时间
	WriteLength   uint64 // 发送的数据长度
	ReadLength    uint64 // 接收的数据长度
	Status        bool   // 状态
	IsStatistics  bool   // 是否开启流量统计,默认为false
	Err           error
}

type ConnectionHandshakeDatabase struct {
	ClientId string
	UserName string
	Password string
	Addr     net.Addr
}
