package rabbitmq

import (
	"errors"
	"fmt"
	"github.com/dc-utils/args"
	"github.com/gofiber/fiber/v2/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

var client *Client

var once sync.Once

func init() {
	enabled := args.GetBool("rabbitmq.enabled", true)
	if !enabled {
		return
	}
	Enable()
}

// Enable
//
//	-uri 例如：amqp://xc2020:123456@localhost:5672/xc_dev
func Enable() {
	uri := args.Get("rabbitmq.uri")
	if uri == "" {
		return
	}
	once.Do(func() {
		maxRetries := args.GetInt("rabbitmq.maxRetries", 2)
		if maxRetries < 0 {
			maxRetries = 0
		}
		consumeLimit := args.GetInt("rabbitmq.consumeLimit", 10)
		if consumeLimit < 1 {
			consumeLimit = 1
		}
		c, _ := NewClient(Config{
			URI:          uri,
			MaxRetries:   maxRetries,
			RetryDelay:   args.GetDuration("rabbitmq.retryDelay", 2*time.Second),
			ConsumeLimit: consumeLimit,
		})
		client = c
	})
}

func Get() *Client {
	return client
}

// Config 连接配置
type Config struct {
	URI          string
	MaxRetries   int
	RetryDelay   time.Duration
	ConsumeLimit int // 最大并发消费者数
}

// Client RabbitMQ客户端
type Client struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	config    Config
	mu        sync.RWMutex
	isClosing bool
}

// NewClient 创建新客户端
func NewClient(config Config) (*Client, error) {
	c := &Client{config: config}
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	return c, nil
}

// 连接RabbitMQ
func (c *Client) connect() error {
	var err error
	for i := 0; i <= c.config.MaxRetries; i++ {
		c.conn, err = amqp.Dial(c.config.URI)
		if err == nil {
			log.Debug("Connected to RabbitMQ")
			return c.createChannel()
		}
		log.Warnf("Connect attempt %d failed: %v", i+1, err)
		time.Sleep(c.config.RetryDelay)
	}
	return fmt.Errorf("max retries exceeded: %w", err)
}

// 创建通道
func (c *Client) createChannel() error {
	var err error
	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel failed: %w", err)
	}
	return nil
}

// Close 关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isClosing = true
	var err error
	if c.channel != nil {
		if closeErr := c.channel.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}

// DeclareExchange 声明交换机
func (c *Client) DeclareExchange(name, kind string, durable bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	if err := c.channel.ExchangeDeclare(
		name,    // name
		kind,    // type
		durable, // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	); err != nil {
		return fmt.Errorf("declare exchange failed: %w", err)
	}
	log.Debugf("Exchange declared: %s", name)
	return nil
}

// DeclareQueue 声明队列
func (c *Client) DeclareQueue(name string, durable bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	if _, err := c.channel.QueueDeclare(
		name,    // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	); err != nil {
		return fmt.Errorf("declare queue failed: %w", err)
	}
	log.Debugf("Queue declared: %s", name)
	return nil
}

// BindQueue 绑定队列到交换机
func (c *Client) BindQueue(queue, exchange, key string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	if err := c.channel.QueueBind(
		queue,    // queue name
		key,      // routing key
		exchange, // exchange
		false,    // no-wait
		nil,      // args
	); err != nil {
		return fmt.Errorf("bind queue failed: %w", err)
	}
	log.Debugf("Queue %s bound to exchange %s with key %s", queue, exchange, key)
	return nil
}

// PublishStr 发布消息
func (c *Client) PublishStr(exchange, key string, body string) error {
	return c.Publish(exchange, key, []byte(body))
}

// Publish 发布消息
func (c *Client) Publish(exchange, key string, body []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	if err := c.channel.Publish(
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}
	return nil
}

// Consume 消费消息
func (c *Client) Consume(queue string, handler func([]byte) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	message, err := c.channel.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("consume failed: %w", err)
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.config.ConsumeLimit)
	go func() {
		for d := range message {
			semaphore <- struct{}{} // 获取令牌
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				defer func() { <-semaphore }() // 释放令牌
				if err := handler(d.Body); err != nil {
					log.Debugf("Handle message error: %v", err)
				}
			}(d)
		}
		wg.Wait()
	}()
	log.Debugf("Consumer started for queue: %s", queue)
	return nil
}

// ConsumeManualAck 消费消息（手动确认模式）
func (c *Client) ConsumeManualAck(queue string, handler func([]byte) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	msgs, err := c.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // 关闭自动确认
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("consume failed: %w", err)
	}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.config.ConsumeLimit)
	go func() {
		for d := range msgs {
			semaphore <- struct{}{} // 获取令牌
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				defer func() { <-semaphore }() // 释放令牌
				if err := handler(d.Body); err != nil {
					log.Debugf("Handle message error: %v, requeuing", err)
					// 处理失败，拒绝消息并重新入队
					if err := d.Nack(false, true); err != nil {
						log.Debugf("Nack failed: %v", err)
					}
					return
				}
				// 处理成功，确认消息
				if err := d.Ack(false); err != nil {
					log.Debugf("Ack failed: %v", err)
				}
			}(d)
		}
		wg.Wait()
	}()
	log.Debugf("Consumer (manual ack) started for queue: %s", queue)
	return nil
}

func (c *Client) DeclareDelayedExchange(name string, durable bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	arguments := amqp.Table{
		"x-delayed-type": "direct", // 支持 direct, topic, fanout 等类型
	}
	if err := c.channel.ExchangeDeclare(
		name,                // name
		"x-delayed-message", // 延时插件类型
		durable,             // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		arguments,           // arguments
	); err != nil {
		return fmt.Errorf("declare delayed exchange failed: %w", err)
	}
	log.Debugf("Delayed exchange declared: %s", name)
	return nil
}

// PublishStrWithDelay 发布延时消息
func (c *Client) PublishStrWithDelay(exchange, key string, body string, delayMillis int) error {
	return c.PublishWithDelay(exchange, key, []byte(body), delayMillis)
}

// PublishWithDelay 发布延时消息
func (c *Client) PublishWithDelay(exchange, key string, body []byte, delayMillis int) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosing {
		return errors.New("client is closing")
	}
	if err := c.channel.Publish(
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"x-delay": delayMillis, // 延时毫秒数
			},
		}); err != nil {
		return fmt.Errorf("publish delayed message failed: %w", err)
	}
	return nil
}
