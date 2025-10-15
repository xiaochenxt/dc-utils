package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
)

// Config 邮件配置信息
type Config struct {
	SMTPHost    string
	SMTPPort    string
	Username    string
	Password    string
	UseTLS      bool // 是否使用 TLS 连接（端口 465）
	UseStartTLS bool // 是否使用 STARTTLS（端口 587）
	SkipVerify  bool // 是否跳过 TLS 证书验证
}

// Email 邮件结构
type Email struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []string
}

// Mailer 邮件发送器
type Mailer struct {
	config Config
}

// NewMailer 创建新的邮件发送器
func NewMailer(config Config) *Mailer {
	return &Mailer{
		config: config,
	}
}

// Send 发送邮件
func (m *Mailer) Send(email Email) error {
	// 验证配置
	if m.config.SMTPHost == "" || m.config.SMTPPort == "" || m.config.Username == "" || m.config.Password == "" {
		return errors.New("SMTP配置不完整")
	}

	// 验证邮件内容
	if email.From == "" {
		email.From = m.config.Username
	}
	if len(email.To) == 0 {
		return errors.New("至少需要一个收件人")
	}

	// 构建邮件消息
	msg, err := buildMessage(email)
	if err != nil {
		return err
	}

	// 构建所有收件人列表
	recipients := make([]string, 0, len(email.To)+len(email.Cc)+len(email.Bcc))
	recipients = append(recipients, email.To...)
	recipients = append(recipients, email.Cc...)
	recipients = append(recipients, email.Bcc...)

	// 连接 SMTP 服务器
	var client *smtp.Client
	var connErr error // 声明单独的错误变量

	serverAddr := m.config.SMTPHost + ":" + m.config.SMTPPort

	if m.config.UseTLS {
		// 使用 TLS 连接（端口 465）
		tlsConfig := &tls.Config{
			ServerName:         m.config.SMTPHost,
			InsecureSkipVerify: m.config.SkipVerify,
		}

		conn, connErr := tls.Dial("tcp", serverAddr, tlsConfig)
		if connErr != nil {
			return connErr
		}

		client, connErr = smtp.NewClient(conn, m.config.SMTPHost)
		if connErr != nil {
			return connErr
		}
	} else {
		// 普通连接
		client, connErr = smtp.Dial(serverAddr)
		if connErr != nil {
			return connErr
		}
		defer client.Quit()

		// 如果启用 STARTTLS
		if m.config.UseStartTLS {
			tlsConfig := &tls.Config{
				ServerName:         m.config.SMTPHost,
				InsecureSkipVerify: m.config.SkipVerify,
			}

			if connErr = client.StartTLS(tlsConfig); connErr != nil {
				return connErr
			}
		}
	}

	// 认证
	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return err
	}

	// 设置发件人和收件人
	if err = client.Mail(email.From); err != nil {
		return err
	}

	for _, addr := range recipients {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

// buildMessage 构建邮件消息
func buildMessage(email Email) ([]byte, error) {
	// 构建邮件头部
	headers := make(textproto.MIMEHeader)
	headers.Set("From", email.From)
	headers.Set("To", joinRecipients(email.To))
	if len(email.Cc) > 0 {
		headers.Set("Cc", joinRecipients(email.Cc))
	}
	headers.Set("Subject", email.Subject)

	// 设置内容类型
	contentType := "text/plain; charset=utf-8"
	if email.IsHTML {
		contentType = "text/html; charset=utf-8"
	}

	// 创建邮件消息
	var msg bytes.Buffer

	// 写入头部
	for k, v := range headers {
		msg.WriteString(k + ": " + v[0] + "\r\n")
	}

	// 处理附件
	if len(email.Attachments) > 0 {
		boundary := "---------------------------1234567890"
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n\r\n")

		// 写入邮件正文
		msg.WriteString("--" + boundary + "\r\n")
		msg.WriteString("Content-Type: " + contentType + "\r\n\r\n")
		msg.WriteString(email.Body + "\r\n\r\n")

		// 添加附件
		for _, filePath := range email.Attachments {
			writeAttachment(filePath, boundary, msg)
		}

		msg.WriteString("--" + boundary + "--\r\n")
	} else {
		// 无附件的简单邮件
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: " + contentType + "\r\n\r\n")
		msg.WriteString(email.Body)
	}

	return msg.Bytes(), nil
}

func writeAttachment(filePath, boundary string, msg bytes.Buffer) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("无法打开附件 %s: %v", filePath, err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("无法读取附件 %s: %v", filePath, err)
		return
	}

	fileName := filepath.Base(filePath)
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: application/octet-stream\r\n")
	msg.WriteString("Content-Disposition: attachment; filename=\"" + fileName + "\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(fileData)))
	base64.StdEncoding.Encode(encoded, fileData)
	msg.Write(encoded)
	msg.WriteString("\r\n\r\n")
}

// joinRecipients 连接收件人列表为字符串
func joinRecipients(recipients []string) string {
	result := ""
	for i, r := range recipients {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}
