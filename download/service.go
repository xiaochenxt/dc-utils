package download

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Service 文件下载服务
type Service struct {
	rate int // 目标速率（字节/秒）
}

// NewService 创建新的下载服务
func NewService(rate int) *Service {
	return &Service{
		rate: rate,
	}
}

// Download 处理文件下载
func (s *Service) Download(c *fiber.Ctx, filepath string) error {
	if s.rate <= 0 {
		return c.Download(filepath)
	}
	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		log.Errorf("打开文件失败: %v\n", err)
		if os.IsNotExist(err) {
			return c.Status(http.StatusNotFound).SendString("文件不存在")
		}
		return c.Status(http.StatusInternalServerError).SendString("服务器错误")
	}
	return s.DownloadFile(c, file)
}

// DownloadFile 处理文件下载
func (s *Service) DownloadFile(c *fiber.Ctx, file *os.File, readerCreator ...io.Reader) error {
	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		err := file.Close()
		if err != nil {
			return err
		}
		log.Errorf("获取文件信息失败: %v\n", err)
		return c.Status(http.StatusInternalServerError).SendString("服务器错误")
	}
	filename := fileInfo.Name()
	log.Debugf("开始处理下载请求: %s\n", filename)
	// 设置响应头
	contentLength := fileInfo.Size()
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Set("Accept-Ranges", "bytes")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// 处理断点续传
	rangeHeader := c.Get("Range")
	if rangeHeader != "" {
		// 解析Range头
		start, end, err := parseRange(rangeHeader, fileInfo.Size())
		if err != nil {
			err := file.Close()
			if err != nil {
				return err
			}
			log.Warnf("解析Range头失败: %v\n", err)
			return c.Status(http.StatusRequestedRangeNotSatisfiable).SendString("无效的Range请求")
		}

		// 部分内容下载
		length := end - start + 1
		c.Set("Content-Length", strconv.FormatInt(length, 10))
		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		c.Status(http.StatusPartialContent)
		log.Debugf("断点续传: 从 %d 到 %d, 长度 %d\n", start, end, length)

		// 定位到起始位置
		if _, err := file.Seek(start, io.SeekStart); err != nil {
			err := file.Close()
			if err != nil {
				return err
			}
			log.Warnf("定位文件失败: %v\n", err)
			return c.Status(http.StatusInternalServerError).SendString("服务器错误")
		}
	} else {
		log.Debugf("完整下载: 文件大小 %d 字节\n", contentLength)
	}
	// 创建限速读取器
	var limitedReader io.Reader
	if len(readerCreator) > 0 {
		limitedReader = readerCreator[0]
	} else {
		limitedReader, err = NewRateLimitedReader(file, s.rate, 8*1024)
		if err != nil {
			err := file.Close()
			if err != nil {
				return err
			}
			log.Warnf("创建限速读取器失败: %v\n", err)
			return c.Status(http.StatusInternalServerError).SendString("创建下载流失败")
		}
	}

	// 使用自定义的closer确保文件在响应完成后关闭
	closer := &responseCloser{
		Reader: limitedReader,
		CloseFunc: func() error {
			log.Debugf("关闭文件: %s\n", filename)
			return file.Close()
		},
	}

	// 设置响应流
	c.Response().SetBodyStream(closer, int(contentLength))

	return nil
}

// responseCloser 实现io.ReadCloser接口
type responseCloser struct {
	io.Reader
	CloseFunc func() error
}

// Close 实现io.Closer接口
func (rc *responseCloser) Close() error {
	return rc.CloseFunc()
}

// parseRange 解析Range头信息
func parseRange(rangeHeader string, fileSize int64) (start, end int64, err error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, errors.New("无效的Range格式")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	ranges := strings.Split(rangeStr, ",")
	if len(ranges) > 1 {
		return 0, 0, errors.New("不支持多个范围")
	}

	parts := strings.Split(ranges[0], "-")
	if len(parts) != 2 {
		return 0, 0, errors.New("无效的Range格式")
	}

	startStr, endStr := parts[0], parts[1]

	if startStr != "" {
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil || start < 0 || start >= fileSize {
			return 0, 0, errors.New("无效的起始位置")
		}
	}

	if endStr != "" {
		end, err = strconv.ParseInt(endStr, 10, 64)
		if err != nil || end < 0 || end >= fileSize {
			return 0, 0, errors.New("无效的结束位置")
		}
	} else {
		end = fileSize - 1
	}

	if startStr == "" && endStr != "" {
		start = fileSize - end
		end = fileSize - 1
	}

	if start > end {
		return 0, 0, errors.New("起始位置不能大于结束位置")
	}

	return start, end, nil
}
