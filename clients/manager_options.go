package clients

import (
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"github.com/qdmc/mqtt_single_proxy/enmu"
)

// ClientManagerOptions   client管理器配置项
type ClientManagerOptions struct {
	TcpPort          uint16                   // tcp监听端口,默认:1883
	IsWebsocket      bool                     // 是否开启websocket,默认:false
	WebsocketPort    uint16                   // websocket监听端口,默认:80
	WebsocketPath    string                   // websocketPath,默认:/websocket
	IsUdp            bool                     // 是否开启udp,默认:false
	UdpPort          uint16                   // udp监听端口,默认:1884
	IsStatistics     bool                     // 是否开启链接数据统计,默认:false
	Handshake        HandshakeHandle          // 握手校验
	ConnectedCb      ConnectedCallback        // 链接回调
	DisConnectCb     DisConnectCallbackHandle // 断开回调
	PacketCb         PacketCallbackHandle     // 报文回调
	WebsocketHandle  WebsocketHandshakeHandle // websocket请求检验
	MaxHandshakeTime int64                    // 握手最大时长(秒),默认:10
	ClientTimeOut    int64                    // 客户端超时(秒),默认:60;客户端在时间内没有报文会断开
}

func (o *ClientManagerOptions) merge(opt *ClientManagerOptions) *ClientManagerOptions {
	if opt == nil {
		return o
	}
	o2 := *opt
	options := &o2
	if options.Handshake == nil {
		options.Handshake = o.Handshake
	}
	if options.ConnectedCb == nil {
		options.ConnectedCb = o.ConnectedCb
	}
	if options.DisConnectCb == nil {
		options.DisConnectCb = o.DisConnectCb
	}
	if options.PacketCb == nil {
		options.PacketCb = o.PacketCb
	}
	if options.WebsocketHandle == nil {
		options.WebsocketHandle = o.WebsocketHandle
	}
	if options.MaxHandshakeTime <= 0 {
		options.MaxHandshakeTime = 10
	}
	return options
}

func newOptions() *ClientManagerOptions {
	return &ClientManagerOptions{
		TcpPort:       1883,
		IsWebsocket:   false,
		WebsocketPort: 80,
		WebsocketPath: "/websocket",
		IsUdp:         false,
		UdpPort:       1884,
		IsStatistics:  false,
		Handshake: func(database clients_dto.ConnectionHandshakeDatabase) enmu.HandshakeResult {
			return enmu.Success
		},
		ConnectedCb:      nil,
		DisConnectCb:     nil,
		PacketCb:         nil,
		WebsocketHandle:  nil,
		MaxHandshakeTime: 10,
	}
}
