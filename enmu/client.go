package enmu

import "errors"

// ClientProtocol  客户端协议
type ClientProtocol string

const (
	TcpProtocol  ClientProtocol = "tcp"
	QuicProtocol ClientProtocol = "quic"
	Websocket    ClientProtocol = "websocket"
)

// HandshakeResult    握手结果
type HandshakeResult byte

const (
	Success                 HandshakeResult = 0x00 // 0x00连接已接受
	ProtocolError           HandshakeResult = 0x01 // 0x01连接已拒绝,不支持的协议版本
	IdError                 HandshakeResult = 0x02 // 0x02连接已拒绝，不合格的客户端标识符
	ServeError              HandshakeResult = 0x03 // 0x03连接已拒绝，服务端不可用
	UserNameOrPasswordError HandshakeResult = 0x04 // 0x04连接已拒绝，无效的用户名或密码
	uUnauthorizedError      HandshakeResult = 0x05 // 0x05连接已拒绝，未授权
)

var ClientDisconnectError = errors.New("client is disconnect")
var ClientHeartTimeoutError = errors.New("client heartbeat is time out")
var PacketEmptyError = errors.New("packet is empty")
var ClientReadConnectionError = errors.New("client read connection error")
var NotFoundClientError = errors.New("not found client")
var NotConnectPacketError = errors.New("this packet is not connectPacket")
var ClienthHandshakeFaild = errors.New("connect handshake failed")
