package rest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// FileValue 表示一个文件值
type FileValue struct {
	FileName   string    // 文件名
	Content    io.Reader // 文件内容
	ContentLen int64     // 文件长度（可选）
}

// Client HTTP客户端封装，可重复使用
type Client struct {
	client  *http.Client
	headers map[string]string
}

// Create 创建HTTP客户端实例，可重复使用
func Create() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 5,
				MaxConnsPerHost:     0,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
		},
		headers: make(map[string]string),
	}
}

func CreateWithGoClient(client *http.Client) *Client {
	client.Jar, _ = cookiejar.New(nil)
	return &Client{
		client:  client,
		headers: make(map[string]string),
	}
}

func CreateNoSSL() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 5,
				MaxConnsPerHost:     0,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		headers: make(map[string]string),
	}
}

// Header 设置默认请求头
func (c *Client) Header(key, value string) *Client {
	c.headers[key] = value
	return c
}

// Headers 设置多个全局默认请求头
func (c *Client) Headers(headers map[string]string) *Client {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

// ClearHeader 清除指定的默认请求头
func (c *Client) ClearHeader(key string) *Client {
	delete(c.headers, key)
	return c
}

// ClearHeaders 清除所有默认请求头
func (c *Client) ClearHeaders() *Client {
	c.headers = make(map[string]string)
	return c
}

// UserAgent 设置UserAgent
func (c *Client) UserAgent(userAgent string) *Client {
	c.headers["User-Agent"] = userAgent
	return c
}

// Request 表示一个HTTP请求
type Request struct {
	client      *Client
	method      string
	url         *url.URL
	headers     map[string]string
	queryParams url.Values
	body        io.Reader
	cookies     []*http.Cookie
}

// newRequest 创建新的请求对象
func (c *Client) newRequest() *Request {
	// 复制客户端的默认请求头
	headers := make(map[string]string)
	for k, v := range c.headers {
		headers[k] = v
	}

	return &Request{
		client:      c,
		headers:     headers,
		queryParams: url.Values{},
	}
}

// Get 发起GET请求
func (c *Client) Get(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodGet
	r.parseUrl(urlStr)
	return r
}

// Post 发起POST请求
func (c *Client) Post(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodPost
	r.parseUrl(urlStr)
	return r
}

// Put 发起PUT请求
func (c *Client) Put(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodPut
	r.parseUrl(urlStr)
	return r
}

// Delete 发起DELETE请求
func (c *Client) Delete(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodDelete
	r.parseUrl(urlStr)
	return r
}

// Options 发起OPTIONS请求
func (c *Client) Options(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodOptions
	r.parseUrl(urlStr)
	return r
}

// Patch 发起PATCH请求
func (c *Client) Patch(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodPatch
	r.parseUrl(urlStr)
	return r
}

// Head 发起HEAD请求
func (c *Client) Head(urlStr string) *Request {
	r := c.newRequest()
	r.method = http.MethodHead
	r.parseUrl(urlStr)
	return r
}

func (r *Request) parseUrl(urlStr string) {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Panicf("URL解析失败: %v", err)
	}
	r.url = u
}

// Header 设置默认请求头
func (r *Request) Header(key, value string) *Request {
	r.headers[key] = value
	return r
}

// Headers 设置多个全局默认请求头
func (r *Request) Headers(headers map[string]string) *Request {
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

// ClearHeader 清除指定的默认请求头
func (r *Request) ClearHeader(key string) *Request {
	delete(r.headers, key)
	return r
}

// ClearHeaders 清除所有默认请求头
func (r *Request) ClearHeaders() *Request {
	r.headers = make(map[string]string)
	return r
}

// UserAgent 设置UserAgent
func (r *Request) UserAgent(userAgent string) *Request {
	r.headers["User-Agent"] = userAgent
	return r
}

// QueryParam 添加查询参数
func (r *Request) QueryParam(key, value string) *Request {
	r.queryParams.Add(key, value)
	return r
}

// QueryParams 添加多个查询参数
func (r *Request) QueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.queryParams.Add(k, v)
	}
	return r
}

// BodyString 设置字符串请求体
func (r *Request) BodyString(body string) *Request {
	r.body = strings.NewReader(body)
	r.headers["Content-Type"] = "text/plain"
	return r
}

// BodyJSON 设置JSON请求体
func (r *Request) BodyJSON(body any) *Request {
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Panicf("JSON序列化失败: %v", err)
	}
	r.body = bytes.NewReader(jsonData)
	r.headers["Content-Type"] = "application/json"
	return r
}

// BodyBytes 设置字节数组请求体
func (r *Request) BodyBytes(body []byte) *Request {
	r.body = bytes.NewReader(body)
	r.headers["Content-Type"] = "application/octet-stream"
	return r
}

// BodyXML 设置XML请求体
func (r *Request) BodyXML(body any) *Request {
	xmlData, err := xml.Marshal(body)
	if err != nil {
		log.Panicf("XML序列化失败: %v", err)
	}
	r.body = bytes.NewReader(xmlData)
	r.headers["Content-Type"] = "application/xml"
	return r
}

// BodyForm 设置表单请求体
func (r *Request) BodyForm(data url.Values) *Request {
	r.body = strings.NewReader(data.Encode())
	r.headers["Content-Type"] = "application/x-www-form-urlencoded"
	return r
}

// BodyMultipart 设置multipart/form-data类型的请求体
func (r *Request) BodyMultipart(fields map[string]any) *Request {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	defer func(bodyWriter *multipart.Writer) {
		_ = bodyWriter.Close()
	}(bodyWriter)
	for key, value := range fields {
		switch v := value.(type) {
		case *multipart.FileHeader:
			// 处理HTTP请求中的文件头
			file, err := v.Open()
			if err != nil {
				log.Panicf("打开文件失败: %v", err)
			}
			fileWriter, _ := bodyWriter.CreateFormFile(key, v.Filename)
			_, _ = io.Copy(fileWriter, file)
			if fileCloser, ok := file.(io.Closer); ok {
				_ = fileCloser.Close()
			}
		case *os.File:
			fileInfo, err := v.Stat()
			if err != nil {
				_ = v.Close()
				log.Panicf("获取文件信息失败: %v", err)
			}
			fileWriter, _ := bodyWriter.CreateFormFile(key, fileInfo.Name())
			_, _ = io.Copy(fileWriter, v)
			_ = v.Close()
		case FileValue:
			fileWriter, _ := bodyWriter.CreateFormFile(key, v.FileName)
			_, _ = io.Copy(fileWriter, v.Content)
			if fileCloser, ok := v.Content.(io.Closer); ok {
				_ = fileCloser.Close()
			}
		case string:
			_ = bodyWriter.WriteField(key, v)
		case int:
			_ = bodyWriter.WriteField(key, strconv.Itoa(v))
		case bool:
			_ = bodyWriter.WriteField(key, strconv.FormatBool(v))
		case []byte:
			fileWriter, _ := bodyWriter.CreateFormFile(key, key)
			reader := bytes.NewReader(v)
			_, _ = io.Copy(fileWriter, reader)
		default:
			_ = bodyWriter.WriteField(key, fmt.Sprintf("%v", v))
		}
	}
	r.body = bodyBuf
	r.headers["Content-Type"] = bodyWriter.FormDataContentType()
	return r
}

// ContentType 手动设置Content-Type
func (r *Request) ContentType(ct string) *Request {
	r.headers["Content-Type"] = ct
	return r
}

// Cookie 添加单个Cookie
func (r *Request) Cookie(cookie *http.Cookie) *Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

// Cookies 添加多个Cookie
func (r *Request) Cookies(cookies []*http.Cookie) *Request {
	r.cookies = append(r.cookies, cookies...)
	return r
}

// CookieKV 快速设置Cookie键值对
func (r *Request) CookieKV(name, value string) *Request {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	}
	r.cookies = append(r.cookies, cookie)
	return r
}

// Retrieve 构建请求但不发送，返回Response对象
func (r *Request) Retrieve() *Response {
	u := r.url
	if len(r.queryParams) > 0 {
		q := u.Query()
		for k, vs := range r.queryParams {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	for _, cookie := range r.cookies {
		r.client.client.Jar.SetCookies(u, []*http.Cookie{cookie})
	}

	var req *http.Request
	if r.body != nil {
		req, _ = http.NewRequest(r.method, r.url.String(), r.body)
	} else {
		req, _ = http.NewRequest(r.method, r.url.String(), nil)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	return &Response{
		request: r,
		req:     req,
	}
}

// Response HTTP响应
type Response struct {
	request   *Request
	req       *http.Request
	resp      *http.Response
	bodyBytes []byte
}

// 发送HTTP请求
func (r *Response) execute() {
	if r.resp != nil {
		return
	}
	resp, err := r.request.client.client.Do(r.req)
	if err != nil {
		log.Panicf("http请求发送失败，%v", err)
	}
	r.resp = resp
}

// 读取响应体并关闭
func (r *Response) readAndClose() {
	if r.resp != nil {
		return
	}
	r.execute()
	defer r.closeBody()
	bodyBytes, err := io.ReadAll(r.resp.Body)
	if err != nil {
		log.Panicf("读取响应体失败: %v", err)
	}
	r.bodyBytes = bodyBytes
}

// 关闭响应体
func (r *Response) closeBody() {
	err := r.resp.Body.Close()
	if err != nil {
		log.Debugf("连接关闭异常: %v", err)
		return
	}
	return
}

// String 返回响应体字符串
func (r *Response) String() string {
	r.readAndClose()
	return string(r.bodyBytes)
}

// JSON 解析响应体为JSON
func (r *Response) JSON(v any) error {
	r.readAndClose()
	if len(r.bodyBytes) == 0 {
		return nil
	}
	return json.Unmarshal(r.bodyBytes, v)
}

// XML 解析响应体为XML
func (r *Response) XML(v any) error {
	r.readAndClose()
	if len(r.bodyBytes) == 0 {
		return nil
	}
	return xml.Unmarshal(r.bodyBytes, v)
}

// Map 解析响应体为Map
func (r *Response) Map() (map[string]any, error) {
	r.readAndClose()
	if len(r.bodyBytes) == 0 {
		return nil, nil
	}
	var m map[string]any
	contentType := r.resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err := json.Unmarshal(r.bodyBytes, &m)
		if err != nil {
			return nil, err
		}
		return m, nil
	} else if strings.HasPrefix(contentType, "application/xml") {
		err := xml.Unmarshal(r.bodyBytes, &m)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	return nil, fmt.Errorf("不支持转换为map，%v", contentType)
}

// Bytes 返回响应体字节数组
func (r *Response) Bytes() []byte {
	r.readAndClose()
	return r.bodyBytes
}

// StatusCode 获取状态码
func (r *Response) StatusCode() int {
	r.readAndClose()
	return r.resp.StatusCode
}

// Header 获取响应头
func (r *Response) Header() http.Header {
	r.readAndClose()
	return r.resp.Header
}

// Cookies 获取响应中的Cookie
func (r *Response) Cookies() []*http.Cookie {
	r.readAndClose()
	return r.resp.Cookies()
}

// Stream 流式处理响应体，每次读取一块数据并传递给回调函数处理
//
// 调用该函数之前或之后不得调用其他Response的操作，如String()，JSON()等，因为响应流只能读取一次
func (r *Response) Stream(readHandler func(bytes []byte), isRead func(statusCode int, header http.Header) bool) {
	r.execute()
	defer r.closeBody()
	if isRead != nil && !isRead(r.resp.StatusCode, r.resp.Header) {
		_, _ = io.Copy(io.Discard, r.resp.Body)
		return
	}
	// 创建固定大小的缓冲区，避免一次性加载整个响应
	buf := make([]byte, 4096)
	for {
		// 从响应体读取一块数据
		n, err := r.resp.Body.Read(buf)
		if n > 0 {
			// 调用处理函数处理当前块
			readHandler(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break // 读取完毕
			}
			log.Panicf("读取响应体失败: %v", err)
		}
	}
}

// StreamJSON 流式解析JSON响应
//
// 调用该函数之前或之后不得调用其他Response的方法，如String()，JSON()等，因为响应流只能读取一次
func (r *Response) StreamJSON(v any) error {
	r.execute()
	defer r.closeBody()
	return json.NewDecoder(r.resp.Body).Decode(v)
}

// StreamXML 流式解析XML响应
//
// 调用该函数之前或之后不得调用其他Response的方法，如String()，JSON()等，因为响应流只能读取一次
func (r *Response) StreamXML(v any) error {
	r.execute()
	defer r.closeBody()
	return xml.NewDecoder(r.resp.Body).Decode(v)
}
