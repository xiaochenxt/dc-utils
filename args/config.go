package args

import (
	"bufio"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strings"
	"time"
)

// 全局配置参数容器
var config = make(map[string]string)

func init() {
	LoadDefaultConfig()
	go func() {
		<-time.After(0 * time.Second)
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			LoadDefaultConfig()
		}
	}()
}

// LoadDefaultConfig 加载默认配置文件
func LoadDefaultConfig() {
	filePath := "application.properties"
	if err := LoadConfig(filePath); err != nil {
		log.Fatalf("%s配置加载失败: %v", filePath, err)
	}
	profiles := GetStrArr("application.profiles.active", nil)
	if profiles != nil && len(profiles) > 0 {
		for _, profile := range profiles {
			fp := "application-" + profile + ".properties"
			if err := LoadConfig(fp); err != nil {
				log.Errorf("%s配置加载失败: %v", fp, err)
			}
		}
	}
}

// LoadConfig 加载并解析properties文件
func LoadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("打开%s配置文件失败: %w", filePath, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	var (
		scanner   = bufio.NewScanner(file)
		multiLine = "" // 用于存储跨行配置的缓冲区
	)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 处理跨行配置
		if strings.HasSuffix(line, "\\") {
			// 移除行尾的反斜杠并添加到多行缓冲区
			multiLine += strings.TrimSpace(strings.TrimSuffix(line, "\\"))
			continue
		} else if multiLine != "" {
			// 如果缓冲区有内容，将当前行添加到缓冲区并处理
			line = multiLine + line
			multiLine = "" // 清空缓冲区
		}
		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Warnf("警告: 无效配置行 - %s", line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}
	return scanner.Err()
}
