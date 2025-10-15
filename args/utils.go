package args

import (
	"github.com/dc-utils/datasize"
	"github.com/dc-utils/dclog"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strconv"
	"strings"
	"time"
)

// 全局命令参数容器
var commands = make(map[string]string)

// 初始化时解析命令行参数
func init() {
	// 解析命令行参数（格式：key=value 或 --key=value）
	for _, arg := range os.Args[1:] {
		// 移除前缀 -- 或 -
		if strings.HasPrefix(arg, "--") {
			arg = strings.TrimPrefix(arg, "--")
		} else if strings.HasPrefix(arg, "-") {
			arg = strings.TrimPrefix(arg, "-")
		} else {
			continue // 不是参数格式，跳过
		}
		// 分割 key=value
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) > 1 {
			commands[parts[0]] = parts[1]
		} else {
			commands[parts[0]] = "true"
		}
	}
	impl := GetStr("log.impl", "")
	if impl == "" {
		log.SetLogger(dclog.DefaultLogger())
	} else if impl == "fiber" {
		log.SetLogger(log.DefaultLogger())
	}
	switch GetStr("log.level", "info") {
	case "trace":
		log.SetLevel(log.LevelTrace)
	case "debug":
		log.SetLevel(log.LevelDebug)
	case "info":
		log.SetLevel(log.LevelInfo)
	case "warn":
		log.SetLevel(log.LevelWarn)
	case "error":
		log.SetLevel(log.LevelError)
	case "fatal":
		log.SetLevel(log.LevelFatal)
	case "panic":
		log.SetLevel(log.LevelPanic)
	default:
	}
}

// Get 获取命令行参数返回字符串，不存在时返回空字符串
func Get(key string) string {
	if val, exists := commands[key]; exists {
		return val
	}
	if val, exists := config[key]; exists {
		return val
	}
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return ""
}

// GetStr 获取命令行参数返回字符串，不存在时返回默认值
func GetStr(key string, defaultValue string) string {
	if val := Get(key); val != "" {
		return val
	}
	return defaultValue
}

// GetStrArr 获取命令行参数返回字符串数组，不存在时返回空字符串数组
func GetStrArr(key string, defaultValue []string) []string {
	if val := Get(key); val != "" {
		values := strings.Split(val, ",")
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}
		return values
	}
	return defaultValue
}

// GetInt 获取命令行参数返回整数值，不存在或转换失败时返回默认值
func GetInt(key string, defaultValue int) int {
	if val := Get(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetUint 获取命令行参数返回无符号整数值，不存在或转换失败时返回默认值
func GetUint(key string, defaultValue uint) uint {
	if val := Get(key); val != "" {
		if i, err := strconv.ParseUint(val, 10, 64); err == nil {
			return uint(i)
		}
	}
	return defaultValue
}

// GetBool 获取命令行参数返回布尔值，不存在或格式错误时返回默认值
func GetBool(key string, defaultValue bool) bool {
	if val := Get(key); val != "" {
		return val == "true" || val == "1" || strings.ToLower(val) == "yes"
	}
	return defaultValue
}

// GetFloat 获取命令行参数返回小数，不存在或转换失败时返回默认值
func GetFloat(key string, defaultValue float64) float64 {
	if val := Get(key); val != "" {
		if i, err := strconv.ParseFloat(val, 10); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetByte 获取命令行参数返回整数值，不存在或转换失败时返回默认值
func GetByte(key string, defaultValue byte) byte {
	if val := Get(key); val != "" {
		if i, err := strconv.ParseUint(val, 10, 8); err == nil {
			return byte(i)
		}
	}
	return defaultValue
}

// GetDuration 获取命令行参数返回时间，不存在或转换失败时返回默认值
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if val := Get(key); val != "" {
		// 尝试使用标准库解析
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
		// 处理包含 "d" 单位的情况
		if strings.HasSuffix(val, "d") {
			// 提取数字部分
			numStr := val[:len(val)-1]
			if days, err := strconv.Atoi(numStr); err == nil {
				// 将天转换为小时，再使用标准库解析
				return time.Duration(days*24) * time.Hour
			}
		}
	}
	return defaultValue
}

// GetDataSize 获取命令行参数返回数据大小，不存在或转换失败时返回默认值
func GetDataSize(key string, defaultValue datasize.DataSize) datasize.DataSize {
	if val := Get(key); val != "" {
		if dataSize, err := datasize.Parse(val); err == nil {
			return dataSize
		}
	}
	return defaultValue
}

// GetDataBytes 获取命令行参数返回数据字节，不存在或转换失败时返回默认值
func GetDataBytes(key string, defaultValue int) int {
	if val := Get(key); val != "" {
		if dataSize, err := datasize.Parse(val); err == nil {
			return dataSize.ToBytes()
		}
	}
	return defaultValue
}
