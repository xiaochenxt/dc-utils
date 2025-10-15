package mem

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GetSystemAvailableMemory 返回系统可用内存的字节数
func GetSystemAvailableMemory() (uint64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxAvailableMemory()
	case "windows":
		return getWindowsAvailableMemory()
	case "darwin":
		return getDarwinAvailableMemory()
	default:
		return 0, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// getLinuxAvailableMemory 通过解析 free 命令的输出来获取可用内存
func getLinuxAvailableMemory() (uint64, error) {
	out, err := exec.Command("free", "-b").Output()
	if err != nil {
		return 0, fmt.Errorf("执行 free 命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("无法解析内存信息: 输出行数不足")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 7 {
		return 0, fmt.Errorf("无法解析内存信息: 字段数不足")
	}

	available, err := strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析可用内存失败: %v", err)
	}

	return available, nil
}

// getWindowsAvailableMemory 通过解析 wmic 命令的输出来获取可用内存
func getWindowsAvailableMemory() (uint64, error) {
	out, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/format:list").Output()
	if err != nil {
		return 0, fmt.Errorf("执行 wmic 命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	var freeMemory uint64

	for _, line := range lines {
		if strings.HasPrefix(line, "FreePhysicalMemory=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			value, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("解析可用内存失败: %v", err)
			}

			// wmic 返回的是 KB，转换为字节
			freeMemory = value * 1024
			break
		}
	}

	if freeMemory == 0 {
		return 0, fmt.Errorf("未找到可用内存信息")
	}

	return freeMemory, nil
}

// getDarwinAvailableMemory 通过解析 vm_stat 命令的输出来获取可用内存
func getDarwinAvailableMemory() (uint64, error) {
	// 获取空闲内存页
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return 0, fmt.Errorf("执行 vm_stat 命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	var freePages uint64

	for _, line := range lines {
		if strings.Contains(line, "free") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			// 移除逗号并解析
			pageStr := strings.ReplaceAll(parts[1], ",", "")
			pages, err := strconv.ParseUint(pageStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("解析空闲页失败: %v", err)
			}

			freePages = pages
			break
		}
	}

	if freePages == 0 {
		return 0, fmt.Errorf("未找到空闲页信息")
	}

	// 每页大小（通常为4096字节）
	pageSize := uint64(4096)
	freeMemory := freePages * pageSize

	return freeMemory, nil
}
