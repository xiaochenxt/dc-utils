package datasize

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"math"
	"strconv"
	"strings"
)

// 定义数据单位常量，类型为 int，可直接参与运算
const (
	B  = 1
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

// DataSize 表示数据大小，本质上是 int 类型
type DataSize int

// FromBytes 创建字节为单位的数据大小
func FromBytes(bytes int) DataSize {
	if bytes < 0 {
		return 0
	}
	return DataSize(bytes)
}

// Parse 解析字符串形式的数据大小，格式错误时返回错误
func Parse(text string) (DataSize, error) {
	if len(text) == 0 {
		return 0, errors.New("empty input")
	}

	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 {
		return 0, errors.New("empty input after trimming")
	}

	// 分离数字和单位
	numEnd := 0
	for numEnd < len(trimmed) && (trimmed[numEnd] >= '0' && trimmed[numEnd] <= '9') {
		numEnd++
	}
	if numEnd == 0 {
		return 0, fmt.Errorf("invalid format: %q", text)
	}

	// 解析数字
	amountStr := trimmed[:numEnd]
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %q", amountStr)
	}
	if amount < 0 {
		return 0, fmt.Errorf("negative value not allowed: %d", amount)
	}

	// 解析单位
	suffix := strings.ToUpper(trimmed[numEnd:])
	multiplier := 1

	switch suffix {
	case "B":
		multiplier = B
	case "KB":
		multiplier = KB
	case "MB":
		multiplier = MB
	case "GB":
		multiplier = GB
	case "TB":
		multiplier = TB
	case "":
		multiplier = B // 默认单位为字节
	default:
		return 0, fmt.Errorf("unknown unit: %q", suffix)
	}

	// 安全乘法，防止溢出
	result := safeMultiply(amount, multiplier)
	return DataSize(result), nil
}

// ToBytes 返回字节数
func (d DataSize) ToBytes() int {
	return int(d)
}

// ToKilobytes 返回千字节数
func (d DataSize) ToKilobytes() int {
	return int(d) / KB
}

// ToMegabytes 返回兆字节数
func (d DataSize) ToMegabytes() int {
	return int(d) / MB
}

// ToGigabytes 返回吉字节数
func (d DataSize) ToGigabytes() int {
	return int(d) / GB
}

// ToTerabytes 返回太字节数
func (d DataSize) ToTerabytes() int {
	return int(d) / TB
}

// CompareTo 比较大小
func (d DataSize) CompareTo(other DataSize) int {
	if d < other {
		return -1
	} else if d > other {
		return 1
	}
	return 0
}

// String 返回可读字符串
func (d DataSize) String() string {
	if d == 0 {
		return "0B"
	}

	bytes := int(d)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// Add 加法操作
func (d DataSize) Add(other DataSize) DataSize {
	return DataSize(safeAdd(int(d), int(other)))
}

// Subtract 减法操作
func (d DataSize) Subtract(other DataSize) DataSize {
	result := int(d) - int(other)
	if result < 0 {
		return 0
	}
	return DataSize(result)
}

// Multiply 乘法操作
func (d DataSize) Multiply(factor int) DataSize {
	return DataSize(safeMultiply(int(d), factor))
}

// Divide 除法操作
func (d DataSize) Divide(divisor int) DataSize {
	if divisor <= 0 {
		return 0
	}
	return DataSize(int(d) / divisor)
}

// safeMultiply 安全乘法，防止溢出
func safeMultiply(a, b int) int {
	if a > math.MaxInt/b {
		log.Errorf("a=%d 乘以 b=%d 超过 int最大值，程序可能出现异常", a, b)
		return math.MaxInt
	}
	return a * b
}

// safeAdd 安全加法，防止溢出
func safeAdd(a, b int) int {
	if a > math.MaxInt-b {
		log.Errorf("a=%d 加 b=%d 超过 int最大值，程序可能出现异常", a, b)
		return math.MaxInt
	}
	return a + b
}
