package sse

import (
	"bufio"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Emitter 负责向客户端发送Server-Sent Events
type Emitter struct {
	w           *bufio.Writer
	ctx         *fiber.Ctx
	nextID      uint64 // 事件ID
	lastEventID string // 存储客户端发送的最后一个事件ID
	onError     func(error)
	onClose     func(error)
	onComplete  func()
	complete    bool
}

// New
//
// 文档地址：https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
//
// 创建一个新的SSE处理器
func New(handler func(*Emitter)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 设置必需的SSE响应头
		c.Set(fiber.HeaderContentType, "text/event-stream")
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Set(fiber.HeaderConnection, "keep-alive")
		c.Set(fiber.HeaderTransferEncoding, "chunked")
		lastEventID := c.Get("Last-Event-ID")
		// 获取底层fasthttp请求上下文
		ctx := c.Context()
		// 设置流写入器
		ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
			e := &Emitter{w: w, ctx: c, lastEventID: lastEventID}
			// 启动心跳机制检测连接状态
			heartbeatTicker := time.NewTicker(5 * time.Second)
			stopHeartbeat := make(chan struct{})
			go func() {
				for {
					select {
					case <-heartbeatTicker.C:
						// 发送心跳注释
						err := e.SendComment("p")
						if err != nil {
							return
						}
					case <-stopHeartbeat:
						return
					}
				}
			}()
			// 调用用户处理函数
			handler(e)
			// 停止心跳
			heartbeatTicker.Stop()
			close(stopHeartbeat)
			// 处理程序完成时调用回调
			if e.onComplete != nil {
				e.complete = true
				e.onComplete()
			}
		})
		return nil
	}
}

func (e *Emitter) send(message string) error {
	_, err := e.w.WriteString(message)
	if err != nil {
		if !e.complete {
			e.complete = true
			if e.onError != nil {
				e.onError(err)
			}
			if e.onClose != nil {
				e.onClose(err)
			}
		}
		return err
	}
	return e.Flush()
}

func (e *Emitter) Flush() error {
	err := e.w.Flush()
	if err != nil {
		if !e.complete {
			e.complete = true
			if e.onError != nil {
				e.onError(err)
			}
			if e.onClose != nil {
				e.onClose(err)
			}
		}
		return err
	}
	return nil
}

func (e *Emitter) sendEvent(event, data, id string, sendId bool) error {
	nData := strings.ReplaceAll(data, "\n", "\ndata:")
	var builder strings.Builder
	if sendId {
		// 如果未提供ID，自动生成一个
		if id == "" {
			id = strconv.FormatUint(atomic.AddUint64(&e.nextID, 1), 10)
		}
		capacity := 32 + len(event) + len(nData) + len(id)
		builder.Grow(capacity)
		builder.WriteString("id:")
		builder.WriteString(id)
		builder.WriteString("\n")
	} else {
		capacity := 32 + len(event) + len(nData)
		builder.Grow(capacity)
	}

	if event != "" {
		builder.WriteString("event:")
		builder.WriteString(event)
		builder.WriteString("\n")
	}

	builder.WriteString("data:")
	builder.WriteString(nData)
	builder.WriteString("\n\n")

	return e.send(builder.String())
}

// SendEvent 发送完整的SSE事件
func (e *Emitter) SendEvent(event, data string) error {
	return e.SendEventWithID(event, data, "")
}

// SendEventNoID 发送不带ID的SSE事件
func (e *Emitter) SendEventNoID(event, data string) error {
	return e.sendEvent(event, data, "", false)
}

// SendEventWithID 发送带ID的SSE事件
func (e *Emitter) SendEventWithID(event, data, id string) error {
	return e.sendEvent(event, data, id, true)
}

// SendData 只发送数据
func (e *Emitter) SendData(data string) error {
	nData := strings.ReplaceAll(data, "\n", "\ndata:")
	capacity := 16 + len(nData)
	var builder strings.Builder
	builder.Grow(capacity)
	builder.WriteString("data:")
	builder.WriteString(nData)
	builder.WriteString("\n\n")
	return e.send(builder.String())
}

// SendRetry 设置客户端重新连接时间
func (e *Emitter) SendRetry(ms uint) error {
	return e.send("retry:" + strconv.FormatUint(uint64(ms), 10) + "\n\n")
}

// SendComment 发送注释（用于保持连接活跃）
func (e *Emitter) SendComment(comment string) error {
	return e.send(":" + comment + "\n\n")
}

func (e *Emitter) Ctx() *fiber.Ctx {
	return e.ctx
}

func (e *Emitter) LastEventID() string {
	return e.lastEventID
}

func (e *Emitter) OnError(handler func(err error)) {
	e.onError = handler
}

func (e *Emitter) OnClose(handler func(err error)) {
	e.onClose = handler
}

func (e *Emitter) OnComplete(handler func()) {
	e.onComplete = handler
}
