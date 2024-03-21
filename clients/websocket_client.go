package clients

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/qdmc/mqtt_packet"
	"github.com/qdmc/mqtt_packet/packets"
	"github.com/qdmc/mqtt_single_proxy/dto/clients_dto"
	"github.com/qdmc/mqtt_single_proxy/enmu"
	"github.com/qdmc/websocket_packet/frame"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type websocketClient struct {
	id                string
	status            bool
	disConnectCb      DisConnectCallbackHandle
	packetCb          PacketCallbackHandle
	connectedNano     int64   // 链接开始时间
	closeNano         int64   // 链接断开时间
	writeLength       *uint64 // 发送的数据长度
	readLength        *uint64 // 接收的数据长度
	isStatistics      bool    // 是否开启流量统计,默认为false
	e                 error
	isNoCb            bool
	conn              net.Conn
	stopChan          chan struct{}
	tr                *time.Timer
	t                 time.Duration
	pt                time.Duration
	ptt               *time.Ticker
	continuationFrame *frame.Frame
	mqttBuf           []byte
}

func (c *websocketClient) GetId() string {
	return c.id
}

func (c *websocketClient) GetDataBase() clients_dto.ConnectionDatabase {
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
func (c *websocketClient) doFrame(f *frame.Frame) error {
	if f == nil {
		return nil
	}
	if f.Opcode == 9 {
		bs, _ := frame.NewPongFrame(f.PayloadData).ToBytes()
		writeLen, err := c.conn.Write(bs)
		if err != nil {
			return err
		}
		if c.isStatistics {
			atomic.AddUint64(c.writeLength, uint64(writeLen))
		}
		return nil
	} else if f.Opcode == 8 {
		c.DisConnect()
		return nil
	} else if f.Opcode == 10 {
		return nil
	} else if f.Opcode == 1 || f.Opcode == 2 {
		c.mqttBuf = append(c.mqttBuf, f.PayloadData...)
		list, lastBs, err := mqtt_packet.ReadStream(c.mqttBuf)
		if err != nil {
			return err
		} else {
			if list != nil && len(list) > 0 {
				for _, p := range list {
					go c.doPacket(p)
				}
			}
			if lastBs != nil && len(lastBs) > 0 {
				c.mqttBuf = lastBs
			} else {
				c.mqttBuf = []byte{}
			}
			return nil
		}
	} else {
		return errors.New("frame type error")
	}
}
func (c *websocketClient) AsyncDoConnection() {
	c.status = true
	var err error
	defer func() {
		c.status = false
		c.closeNano = time.Now().UnixNano()
		c.doDisconnect(err)
	}()
	c.tr = time.NewTimer(c.t)
	c.ptt = time.NewTicker(c.pt)
	for {
		select {
		case <-c.stopChan:
			err = nil
			return
		case <-c.ptt.C:
			go c.doPing()
			continue
		case <-c.tr.C:
			err = enmu.ClientHeartTimeoutError
			return
		default:
			readLen, f, code := frame.ReadOnceFrame(c.conn)
			if code != frame.CloseNormalClosure {
				err = enmu.ClientReadConnectionError
				return
			}
			c.tr.Reset(c.t)
			if c.isStatistics {
				atomic.AddUint64(c.readLength, uint64(readLen))
			}
			if f.Fin == 0 {
				if c.continuationFrame == nil {
					c.continuationFrame = f
				} else {
					c.continuationFrame.PayloadData = append(c.continuationFrame.PayloadData, f.PayloadData...)
				}
			} else {
				// 这里合并分包,并弹出
				if c.continuationFrame != nil {
					composeFrame := new(frame.Frame)
					composeFrame.SetOpcode(f.Opcode)
					composeFrame.SetPayload(append(c.continuationFrame.PayloadData, f.PayloadData...))
					c.continuationFrame = nil
					err = c.doFrame(composeFrame)
				} else {
					err = c.doFrame(f)
				}
				if err != nil {
					return
				}
			}
			continue
		}
	}
}
func (c *websocketClient) doPing() {
	if c.status {
		frameBytes, _ := frame.NewPingFrame([]byte("hello")).ToBytes()
		writeLen, err := c.conn.Write(frameBytes)
		if err != nil {
			atomic.AddUint64(c.writeLength, uint64(writeLen))
		}
	}
}
func (c *websocketClient) SetTimeOut(t int64) {
	if !c.status {
		if t >= 10 && t <= 300 {
			c.t = time.Duration(t) * time.Second
			c.pt = time.Duration(t-5) * time.Second
		}
	}
}
func (c *websocketClient) SetStatistics(b bool) {
	if !c.status {
		c.isStatistics = b
	}
}

func (c *websocketClient) SetPacketHandle(handle PacketCallbackHandle) {
	if !c.status {
		c.packetCb = handle
	}
}

func (c *websocketClient) SetDisConnectCallback(handle DisConnectCallbackHandle) {
	if !c.status {
		c.disConnectCb = handle
	}
}

func (c *websocketClient) DisConnect(isNoCb ...bool) {
	if isNoCb != nil && len(isNoCb) == 1 && isNoCb[0] {
		c.isNoCb = true
	}
	close(c.stopChan)
}

func (c *websocketClient) GetProtocol() enmu.ClientProtocol {
	return enmu.TcpProtocol
}
func (c *websocketClient) WritePacketOnce(p mqtt_packet.ControlPacketInterface) (int64, error) {
	if p == nil {
		return 0, enmu.PacketEmptyError
	}
	if c.status {
		mqBuf := bytes.NewBuffer([]byte{})
		_, err := p.Write(mqBuf)
		if err != nil {
			return 0, err
		}
		bs, err := frame.AutoBinaryFramesBytes(mqBuf.Bytes())
		if err != nil {
			return 0, err
		}
		l, err := c.conn.Write(bs)
		if err != nil {
			return 0, err
		}
		return int64(l), nil
	} else {
		return 0, enmu.ClientDisconnectError
	}
}
func (c *websocketClient) doPacket(p mqtt_packet.ControlPacketInterface) {
	if p == nil || c.packetCb == nil {
		return
	}
	go c.packetCb(c.id, p)
}

func (c *websocketClient) doDisconnect(err error) {
	if err != nil {
		c.e = err
	}
	if c.tr != nil {
		c.tr.Stop()
	}
	if c.ptt != nil {
		c.ptt.Stop()
	}
	if !c.isNoCb && c.disConnectCb != nil {
		db := c.GetDataBase()
		go c.disConnectCb(&db)
	}
}
func newWebsocketClient(id string, c net.Conn) *websocketClient {
	var rl, wl uint64
	return &websocketClient{
		id:                id,
		status:            false,
		disConnectCb:      nil,
		packetCb:          nil,
		connectedNano:     time.Now().UnixNano(),
		closeNano:         0,
		writeLength:       &wl,
		readLength:        &rl,
		isStatistics:      false,
		e:                 nil,
		isNoCb:            false,
		conn:              nil,
		stopChan:          make(chan struct{}, 1),
		tr:                nil,
		t:                 time.Duration(60) * time.Second,
		pt:                time.Duration(55) * time.Second,
		ptt:               nil,
		continuationFrame: nil,
		mqttBuf:           nil,
	}
}
func handshakeWebsocket(c net.Conn, handle HandshakeHandle, handshakeTime int64) (*websocketClient, error) {
	var err error
	if handshakeTime <= 0 {
		handshakeTime = 10
	}
	err = c.SetDeadline(time.Now().Add(time.Duration(handshakeTime) * time.Second))
	if err != nil {
		return nil, err
	}
	_, f, code := frame.ReadOnceFrame(c)
	if code != frame.CloseNormalClosure {
		return nil, enmu.ClientReadConnectionError
	}
	_, p, err := mqtt_packet.ReadOnce(bytes.NewBuffer(f.PayloadData))
	if err != nil {
		return nil, err
	}
	packet, ok := p.(*packets.ConnectPacket)
	if !ok {
		return nil, enmu.NotConnectPacketError
	}
	if handle != nil {
		hd := clients_dto.ConnectionHandshakeDatabase{
			ClientId: packet.ClientIdentifier,
			UserName: packet.Username,
			Password: string(packet.Password),
			Addr:     c.RemoteAddr(),
		}
		res := handle(hd)
		if res != enmu.Success {
			return nil, enmu.ClienthHandshakeFaild
		}
	}
	err = c.SetDeadline(time.Time{})
	if err != nil {
		return nil, err
	}
	return newWebsocketClient(packet.ClientIdentifier, c), nil
}

// websocketUpgradeHandler      websocket校验握手
func websocketUpgradeHandler(req *http.Request, w http.ResponseWriter, otherHandle func(req *http.Request) error) (conn net.Conn, err error) {
	err = defaultUpgradeCheck(req)
	if err != nil {
		return
	}
	if otherHandle != nil {
		err = otherHandle(req)
		if err != nil {
			return
		}
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("this ResponseWriter is not Hijacker")
	}
	conn, _, err = hj.Hijack()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("HijackErr: %s", err.Error()))
	}
	return conn, nil
}

func defaultUpgradeCheck(r *http.Request) error {
	if r.Method != http.MethodGet {
		return errors.New("bad method")
	}
	if !checkHttpHeaderKeyVale(r.Header, "Connection", "upgrade") {
		return errors.New("missing or bad upgrade")
	}
	if !checkHttpHeaderKeyVale(r.Header, "Upgrade", "websocket") {
		return errors.New("missing or bad WebSocket-Protocol")
	}
	if !checkHttpHeaderKeyVale(r.Header, "Sec-Websocket-Version", "13") {
		return errors.New("bad protocol version")
	}
	if !checkSecWebsocketKey(r.Header) {
		return errors.New("bad 'Sec-WebSocket-Key'")
	}
	return nil
}

func checkHttpHeaderValue(h http.Header, key string) ([]string, bool) {
	if key == "" {
		return nil, false
	}
	if val, ok := h[key]; ok {
		return val, true
	} else {
		return nil, false
	}
}

func checkHttpHeaderKeyVale(h http.Header, key, value string) bool {
	if vals, ok := checkHttpHeaderValue(h, key); ok {
		valueStr := strings.ToLower(value)
		for _, val := range vals {
			if strings.Index(strings.ToLower(val), valueStr) != -1 {
				return true
			}
		}
		return false
	} else {
		return false
	}
}

// checkSecWebsocketKey   校验Sec-Websocket-Key
func checkSecWebsocketKey(h http.Header) bool {
	key := h.Get("Sec-Websocket-Key")
	if key == "" {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	return err == nil && len(decoded) == 16
}

// makeServerHandshakeBytes    生成服务端回复的报文
func makeServerHandshakeBytes(req *http.Request) []byte {
	key := req.Header.Get("Sec-Websocket-Key")
	var p []byte
	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, computeAcceptKey(key)...)
	p = append(p, "\r\n"...)
	p = append(p, "\r\n"...)
	return p
}

// httpResponseError    Response回复错误,在拆解Response前使用
func httpResponseError(w http.ResponseWriter, status int, err error) {
	errStr := http.StatusText(status)
	if err != nil && err.Error() != "" {
		errStr = err.Error()
	}
	http.Error(w, errStr, status)
}

// generateChallengeKey   生成随机的websocketKey
func generateChallengeKey() (string, error) {
	p := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, p); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(p), nil
}

// computeAcceptKey     计算websocket的key
func computeAcceptKey(key string) string {
	h := sha1.New() //#nosec G401 -- (CWE-326) https://datatracker.ietf.org/doc/html/rfc6455#page-54
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
