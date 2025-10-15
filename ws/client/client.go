package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
	"net/http"
	"sync/atomic"
	"time"
)

// Client 封装WebSocket客户端功能
type Client struct {
	url              string
	headers          http.Header
	reconnect        bool
	reconnectInt     time.Duration
	maxRetries       int
	isReconnect      func(websocket.CloseError) bool
	enabledHeartbeat bool
	// 回调函数
	onConnect       func(*websocket.Conn)
	onDisconnect    func(error)
	onTextMessage   func(string)
	onBinaryMessage func([]byte)
	onError         func(error)
}

// Create 创建新的WebSocket客户端，默认断开连接会自动重连
func Create(url string) *Client {
	return &Client{
		url:          url,
		reconnect:    true,
		reconnectInt: 5 * time.Second,
		maxRetries:   5, // 默认最大重试5次
		headers:      make(http.Header),
	}
}

// Headers 设置HTTP请求头
func (c *Client) Headers(headers http.Header) *Client {
	c.headers = headers
	return c
}

// Reconnect 设置重连
func (c *Client) Reconnect(interval time.Duration, maxRetries int, isReconnect func(closeError websocket.CloseError) bool) *Client {
	c.reconnectInt = interval
	c.maxRetries = maxRetries
	c.isReconnect = isReconnect
	return c
}

func (c *Client) EnableHeartbeat() *Client {
	c.enabledHeartbeat = true
	return c
}

// DisableReconnect 禁用重连
func (c *Client) DisableReconnect() *Client {
	c.reconnect = false
	return c
}

// OnConnect 设置连接建立回调
func (c *Client) OnConnect(callback func(*websocket.Conn)) *Client {
	c.onConnect = callback
	return c
}

// OnDisconnect 设置断开连接回调
func (c *Client) OnDisconnect(callback func(error)) *Client {
	c.onDisconnect = callback
	return c
}

// OnError 设置错误回调
func (c *Client) OnError(callback func(error)) *Client {
	c.onError = callback
	return c
}

// OnTextMessage 设置文本消息回调
func (c *Client) OnTextMessage(callback func(message string)) *Client {
	c.onTextMessage = callback
	return c
}

// OnBinaryMessage 设置二进制消息回调
func (c *Client) OnBinaryMessage(callback func(message []byte)) *Client {
	c.onBinaryMessage = callback
	return c
}

// Connect 连接到WebSocket服务器，并返回Session
func (c *Client) Connect() (*Session, error) {
	session := &Session{
		client:          c,
		conn:            nil,
		closed:          atomic.Bool{},
		retryCount:      0,
		heartbeatTicker: nil,
		heartbeatDone:   make(chan struct{}),
		heartbeatActive: atomic.Bool{},
	}
	session.closed.Store(true)
	session.heartbeatActive.Store(false)
	err := session.connect()
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Session 表示一个WebSocket会话
type Session struct {
	client          *Client
	conn            *websocket.Conn
	closed          atomic.Bool
	retryCount      int
	heartbeatTicker *time.Ticker
	heartbeatDone   chan struct{}
	heartbeatActive atomic.Bool
}

// connect 建立WebSocket连接，返回nil代表连接成功，返回error则连接失败
func (s *Session) connect() error {
	s.retryCount = 0

	var err error
	var resp *http.Response

	// 基础重试间隔
	baseInterval := s.client.reconnectInt
	ctx := context.Background()
	for {
		s.conn, resp, err = websocket.DefaultDialer.Dial(s.client.url, s.client.headers)
		if err == nil {
			if resp != nil {
				_ = resp.Body.Close()
			}
			s.closed.Store(false)
			s.retryCount = 0
			log.Debugf("WebSocket连接成功: %s", s.client.url)
			if s.client.onConnect != nil {
				s.client.onConnect(s.conn)
			}
			if s.client.enabledHeartbeat {
				s.startHeartbeat()
			}
			// 启动消息处理
			go s.readMessages(ctx)
			return nil
		}

		if !s.client.reconnect {
			return fmt.Errorf("WebSocket连接失败，%w", err)
		}

		// 计算当前重试间隔（指数退避）
		retryInterval := baseInterval * time.Duration(1<<s.retryCount)
		if retryInterval > 2*time.Minute { // 最大间隔2分钟
			retryInterval = 2 * time.Minute
		}

		s.retryCount++
		if s.client.maxRetries > 0 && s.retryCount > s.client.maxRetries {
			return fmt.Errorf("WebSocket连接失败，超过最大重试次数(%d): %w", s.client.maxRetries, err)
		}

		log.Debugf("WebSocket连接失败 (尝试 %d/%d): %v，将在 %v 后重试",
			s.retryCount, s.client.maxRetries, err, retryInterval)

		time.Sleep(retryInterval)
	}
}

// Send 发送消息
func (s *Session) Send(messageType int, data []byte) {
	if s.closed.Load() {
		return
	}
	err := s.conn.WriteMessage(messageType, data)
	if err != nil {
		log.Debugf("WebSocket发送消息失败：%v", err)
		return
	}
}

// SendText 发送文本消息
func (s *Session) SendText(text string) {
	s.Send(websocket.TextMessage, []byte(text))
}

// SendBytes 发送二进制消息
func (s *Session) SendBytes(bytes []byte) {
	s.Send(websocket.BinaryMessage, bytes)
}

// SendJSON 发送JSON消息
func (s *Session) SendJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Warnf("WebSocket发送JSON消息失败：%v", err)
		return
	}
	s.Send(websocket.TextMessage, data)
}

// startHeartbeat 启动心跳，每29秒发送一次空字节数组的Ping消息
func (s *Session) startHeartbeat() *Session {
	if s.heartbeatActive.Load() {
		return s
	}
	s.heartbeatActive.Store(true)
	if s.heartbeatTicker == nil {
		s.heartbeatTicker = time.NewTicker(29 * time.Second)
	} else {
		s.heartbeatTicker.Reset(29 * time.Second)
	}
	// 如果通道为nil，则创建新的通道
	select {
	case <-s.heartbeatDone:
		s.heartbeatDone = make(chan struct{})
	default:
		// 通道未关闭，无需操作
	}
	log.Debug("WebSocket启动心跳")
	go func() {
		for {
			select {
			case <-s.heartbeatDone:
				return
			case <-s.heartbeatTicker.C:
				s.sendPong()
			}
		}
	}()
	return s
}

// stopHeartbeat 停止心跳
func (s *Session) stopHeartbeat() {
	if !s.heartbeatActive.Load() {
		return
	}
	s.heartbeatActive.Store(false)
	if s.heartbeatTicker != nil {
		s.heartbeatTicker.Stop()
	}
	close(s.heartbeatDone)
	log.Debug("WebSocket停止心跳")
}

// sendPong 发送Pong消息
func (s *Session) sendPong() {
	if s.closed.Load() {
		return
	}
	_ = s.conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(1*time.Second))
}

// Close 关闭连接
func (s *Session) Close() {
	// 停止心跳
	s.stopHeartbeat()
	if s.closed.Load() {
		return
	}
	s.closed.Store(true)

	// 发送关闭帧
	_ = s.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 关闭底层连接
	_ = s.conn.Close()
}

// readMessages 读取消息循环
func (s *Session) readMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 读取消息
		messageType, message, err := s.conn.ReadMessage()
		if err != nil {
			s.handleReadError(err)
			return
		}
		// 处理消息
		if messageType == websocket.TextMessage {
			if s.client.onTextMessage != nil {
				s.client.onTextMessage(string(message))
			} else {
				log.Debugf("WebSocket收到文本消息：%v", string(message))
			}
		} else if messageType == websocket.BinaryMessage {
			if s.client.onBinaryMessage != nil {
				s.client.onBinaryMessage(message)
			}
		}
	}
}

// handleReadError 处理读取错误并尝试重连
func (s *Session) handleReadError(err error) {
	s.closed.Store(true)
	_ = s.conn.Close()
	s.stopHeartbeat()
	// 记录错误
	if s.client.onError != nil {
		s.client.onError(err)
	}

	// 触发断开连接回调
	if s.client.onDisconnect != nil {
		s.client.onDisconnect(err)
	}

	// 尝试重连
	if s.client.reconnect {
		var closeErrorType *websocket.CloseError
		if errors.As(err, &closeErrorType) {
			closeErr := *err.(*websocket.CloseError)
			if s.client.isReconnect == nil {
				if !(closeErr.Code == websocket.CloseGoingAway || closeErr.Code == websocket.CloseInternalServerErr ||
					closeErr.Code == websocket.CloseServiceRestart || closeErr.Code == websocket.CloseTryAgainLater ||
					closeErr.Code == websocket.CloseAbnormalClosure) {
					return
				}
			} else {
				if !s.client.isReconnect(closeErr) {
					return
				}
			}
			go func() {
				log.Debugf("WebSocket连接断开，尝试重连: %v", err)
				err := s.connect()
				if err != nil {
					log.Error(err)
				}
			}()
		}
	}
}
