package download

import (
	"io"
	"os"
	"time"
)

// delayRateLimitedReader 延迟限速读取器
type delayRateLimitedReader struct {
	r                *rateLimitedReader
	newRateLimit     int64
	newRateLimitTime time.Time
	isChange         bool
}

// NewDelayRateLimitedReader 创建新的延迟限速读取器
func NewDelayRateLimitedReader(file *os.File, rateLimit, bufferSize int, newRateLimit int64, newRateLimitTime time.Time) (io.Reader, error) {
	reader, err := NewRateLimitedReader(file, rateLimit, bufferSize)
	if err != nil {
		return nil, err
	}
	return &delayRateLimitedReader{
		r:                reader.(*rateLimitedReader),
		newRateLimit:     newRateLimit,
		newRateLimitTime: newRateLimitTime,
	}, nil
}

func (rlr *delayRateLimitedReader) Read(p []byte) (n int, err error) {
	if !rlr.isChange && time.Now().After(rlr.newRateLimitTime) {
		rlr.r.ChangeRateLimit(rlr.newRateLimit)
		rlr.isChange = true
	}
	return rlr.r.Read(p)
}
