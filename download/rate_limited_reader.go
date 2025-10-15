package download

import (
	"bufio"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

// RateLimitedReader 限速读取器
type rateLimitedReader struct {
	file               *os.File
	rateBytesPerSecond int64      // 目标速率（字节/秒）
	startTime          time.Time  // 开始时间
	bytesRead          int64      // 已读取字节数
	mu                 sync.Mutex // 互斥锁
}

// applyRateLimit 应用速率限制
func (rlr *rateLimitedReader) applyRateLimit(bytesToRead int64) {
	// 计算已过去的时间（毫秒）
	elapsedTime := time.Since(rlr.startTime).Milliseconds()
	// 计算读取这些字节应该花费的预期时间（毫秒）
	expectedTime := int64(math.Floor(float64(rlr.bytesRead+bytesToRead) * 1000.0 / float64(rlr.rateBytesPerSecond)))
	// 如果已用时间小于预期时间，需要等待
	if elapsedTime < expectedTime {
		sleepTime := expectedTime - elapsedTime
		// 针对Go协程使用较长的休眠时间
		if sleepTime > 0 {
			maxSleep := int64(1000) // 最大休眠1000ms
			if sleepTime > maxSleep {
				sleepTime = maxSleep
			}
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		}
	}
}

func (rlr *rateLimitedReader) ChangeRateLimit(rateLimit int64) {
	rlr.mu.Lock()
	defer rlr.mu.Unlock()
	rlr.rateBytesPerSecond = rateLimit
	rlr.startTime = time.Now()
	rlr.bytesRead = 0
}

// alignTo4K 将大小向下对齐为4K的倍数
func alignTo4K(size int) int {
	if size <= 4096 {
		return 4096
	}
	return (size / 4096) * 4096
}

// NewRateLimitedReader 创建新的限速读取器
func NewRateLimitedReader(file *os.File, rateLimit, bufferSize int) (io.Reader, error) {
	// 设置最小速率限制为20KB/s
	if rateLimit < 20*1024 {
		rateLimit = 20 * 1024
	} else {
		rateLimit = alignTo4K(rateLimit)
	}
	reader := &rateLimitedReader{
		file:               file,
		rateBytesPerSecond: int64(rateLimit),
		startTime:          time.Now(),
	}
	if bufferSize > 0 {
		return bufio.NewReaderSize(reader, alignTo4K(bufferSize)), nil
	}
	return reader, nil
}

// Read 实现io.Reader接口
func (rlr *rateLimitedReader) Read(p []byte) (n int, err error) {
	rlr.mu.Lock()
	defer rlr.mu.Unlock()
	// 计算本次可以读取的最大字节数
	bytesToRead := int64(len(p))
	// 应用速率限制
	rlr.applyRateLimit(bytesToRead)
	// 读取数据
	n, err = rlr.file.Read(p)
	if n > 0 {
		rlr.bytesRead += int64(n)
	}
	return n, err
}
