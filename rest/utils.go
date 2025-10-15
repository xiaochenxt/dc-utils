package rest

var client = Create()

// Get 发起GET请求
func Get(urlStr string) *Request {
	return client.Get(urlStr)
}

// Post 发起POST请求
func Post(urlStr string) *Request {
	return client.Post(urlStr)
}

// Put 发起PUT请求
func Put(urlStr string) *Request {
	return client.Put(urlStr)
}

// Delete 发起DELETE请求
func Delete(urlStr string) *Request {
	return client.Delete(urlStr)
}

// Options 发起OPTIONS请求
func Options(urlStr string) *Request {
	return client.Options(urlStr)
}

// Patch 发起PATCH请求
func Patch(urlStr string) *Request {
	return client.Patch(urlStr)
}

// Head 发起HEAD请求
func Head(urlStr string) *Request {
	return client.Head(urlStr)
}
