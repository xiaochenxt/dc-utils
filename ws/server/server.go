package server

import (
	"container/list"
	"github.com/gofiber/contrib/websocket"
	"log"
	"sync"
)

// Session 表示一个连接会话
type Session struct {
	UserID    string          // 用户唯一标识
	Conn      *websocket.Conn // WebSocket 连接
	Data      map[string]any  // 自定义用户数据
	CreatedAt int64           // 连接创建时间戳（毫秒）
	mu        sync.Mutex      // 保护连接的并发访问
}

// SessionManager 管理所有客户端连接会话
type SessionManager struct {
	sessions        map[string]*list.List // userId -> 客户端连接会话列表
	maxConnsPerUser int                   // 每个用户的最大连接会话数
	mu              sync.RWMutex          // 读写锁，保护 clients 映射
}

// NewSessionManager 创建一个新的连接会话管理器
func NewSessionManager(maxConnsPerUser int) *SessionManager {
	return &SessionManager{
		sessions:        make(map[string]*list.List),
		maxConnsPerUser: maxConnsPerUser,
	}
}

// AddSession 添加一个连接会话
func (m *SessionManager) AddSession(client *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取或创建用户的连接列表
	if _, ok := m.sessions[client.UserID]; !ok {
		m.sessions[client.UserID] = list.New()
	}

	connList := m.sessions[client.UserID]

	// 检查是否超过最大连接数
	if connList.Len() >= m.maxConnsPerUser {
		// 关闭最老的连接（列表头部）
		oldestElem := connList.Front()
		if oldestElem != nil {
			oldestClient := oldestElem.Value.(*Session)
			log.Printf("用户 %s 连接数达到上限(%d)，关闭最早的连接",
				client.UserID, m.maxConnsPerUser)
			oldestClient.Close()
			connList.Remove(oldestElem)
		}
	}

	// 添加新连接会话到列表尾部（最新）
	connList.PushBack(client)
	log.Printf("用户 %s 连接成功，当前连接数: %d",
		client.UserID, connList.Len())
}

// RemoveSession 注销一个客户端
func (m *SessionManager) RemoveSession(session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if connList, ok := m.sessions[session.UserID]; ok {
		// 查找并移除该客户端
		for elem := connList.Front(); elem != nil; elem = elem.Next() {
			if elem.Value.(*Session) == session {
				session.Close()
				connList.Remove(elem)
				log.Printf("用户 %s 已注销，当前连接数: %d",
					session.UserID, connList.Len())
				break
			}
		}

		// 如果列表为空，删除用户条目
		if connList.Len() == 0 {
			delete(m.sessions, session.UserID)
		}
	}
}

// GetSessions 获取指定用户的所有连接会话
func (m *SessionManager) GetSessions(userId string) []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if connList, ok := m.sessions[userId]; ok {
		clients := make([]*Session, 0, connList.Len())
		for elem := connList.Front(); elem != nil; elem = elem.Next() {
			clients = append(clients, elem.Value.(*Session))
		}
		return clients
	}

	return nil
}

// SendToUser 向指定用户的所有连接会话发送消息
func (m *SessionManager) SendToUser(userId string, messageType int, payload []byte) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if connList, ok := m.sessions[userId]; ok {
		sentCount := 0
		for elem := connList.Front(); elem != nil; elem = elem.Next() {
			client := elem.Value.(*Session)
			if client.Send(messageType, payload) {
				sentCount++
			}
		}
		return sentCount
	}

	log.Printf("用户 %s 没有活跃连接", userId)
	return 0
}

// Broadcast 向所有用户广播消息
func (m *SessionManager) Broadcast(messageType int, payload []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for userId, connList := range m.sessions {
		go func(id string, list *list.List) {
			for elem := list.Front(); elem != nil; elem = elem.Next() {
				client := elem.Value.(*Session)
				client.Send(messageType, payload)
			}
		}(userId, connList)
	}
}

// Close 关闭所有连接会话
func (m *SessionManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for userId, connList := range m.sessions {
		for elem := connList.Front(); elem != nil; elem = elem.Next() {
			client := elem.Value.(*Session)
			client.Close()
		}
		delete(m.sessions, userId)
	}

	log.Printf("所有连接会话已关闭")
}

// Send 向客户端发送消息
func (c *Session) Send(messageType int, payload []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn == nil {
		log.Printf("用户 %s 的连接会话已关闭", c.UserID)
		return false
	}

	err := c.Conn.WriteMessage(messageType, payload)
	if err != nil {
		log.Printf("发送消息到用户 %s 失败: %v", c.UserID, err)
		c.Close()
		return false
	}

	return true
}

// Close 关闭客户端连接
func (c *Session) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn != nil {
		_ = c.Conn.Close()
		c.Conn = nil
	}
}
