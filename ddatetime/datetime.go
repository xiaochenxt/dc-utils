package ddatetime

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LocalDateTime 表示不带时区的日期时间，支持 nil 值
type LocalDateTime struct {
	time *time.Time
}

// 零值时间（Go的零时间）
var zeroTime = time.Time{}

// Now 返回当前日期时间
func Now() LocalDateTime {
	t := time.Now()
	return LocalDateTime{time: &t}
}

// Of 创建指定的日期时间
func Of(year int, month time.Month, day, hour, minute, second int) LocalDateTime {
	t := time.Date(year, month, day, hour, minute, second, 0, time.Local)
	return LocalDateTime{time: &t}
}

// Parse 解析字符串为 LocalDateTime，支持 ISO 格式和自定义格式
func Parse(str string) (LocalDateTime, error) {
	if str == "" {
		return LocalDateTime{}, nil
	}
	// 尝试带毫秒的格式
	t, err := time.Parse("2006-01-02 15:04:05.000000", str)
	if err == nil {
		return LocalDateTime{time: &t}, nil
	}
	// 尝试 ISO 格式
	t, err = time.Parse("2006-01-02T15:04:05", str)
	if err == nil {
		return LocalDateTime{time: &t}, nil
	}
	// 尝试常见的 "yyyy-MM-dd HH:mm:ss" 格式
	t, err = time.Parse("2006-01-02 15:04:05", str)
	if err == nil {
		return LocalDateTime{time: &t}, nil
	}
	return LocalDateTime{}, errors.New("不支持的日期时间格式")
}

// IsZero 检查时间是否为零值（空时间）
func (ldt LocalDateTime) IsZero() bool {
	return ldt.time == nil || *ldt.time == zeroTime
}

// Format 格式化为字符串
func (ldt LocalDateTime) Format(format string) string {
	if ldt.IsZero() {
		return ""
	}
	goFormat := replaceJavaFormatSymbols(format)
	return ldt.time.Format(goFormat)
}

// PlusYears 添加指定年数
func (ldt LocalDateTime) PlusYears(years int) LocalDateTime {
	if ldt.IsZero() {
		return LocalDateTime{}
	}
	t := ldt.time.AddDate(years, 0, 0)
	return LocalDateTime{time: &t}
}

// PlusMonths 添加指定月数
func (ldt LocalDateTime) PlusMonths(months int) LocalDateTime {
	if ldt.IsZero() {
		return LocalDateTime{}
	}
	t := ldt.time.AddDate(0, months, 0)
	return LocalDateTime{time: &t}
}

// PlusDays 添加指定天数
func (ldt LocalDateTime) PlusDays(days int) LocalDateTime {
	if ldt.IsZero() {
		return LocalDateTime{}
	}
	t := ldt.time.AddDate(0, 0, days)
	return LocalDateTime{time: &t}
}

// PlusHours 添加指定小时数
func (ldt LocalDateTime) PlusHours(hours int) LocalDateTime {
	if ldt.IsZero() {
		return LocalDateTime{}
	}
	t := ldt.time.Add(time.Duration(hours) * time.Hour)
	return LocalDateTime{time: &t}
}

// PlusMinutes 添加指定分钟数
func (ldt LocalDateTime) PlusMinutes(minutes int) LocalDateTime {
	if ldt.IsZero() {
		return LocalDateTime{}
	}
	t := ldt.time.Add(time.Duration(minutes) * time.Minute)
	return LocalDateTime{time: &t}
}

// PlusSeconds 添加指定秒数
func (ldt LocalDateTime) PlusSeconds(seconds int) LocalDateTime {
	if ldt.IsZero() {
		return LocalDateTime{}
	}
	t := ldt.time.Add(time.Duration(seconds) * time.Second)
	return LocalDateTime{time: &t}
}

// Year 返回年份
func (ldt LocalDateTime) Year() int {
	if ldt.IsZero() {
		return 0
	}
	return ldt.time.Year()
}

// Month 返回月份（1-12）
func (ldt LocalDateTime) Month() int {
	if ldt.IsZero() {
		return 0
	}
	return int(ldt.time.Month())
}

// Day 返回日期（1-31）
func (ldt LocalDateTime) Day() int {
	if ldt.IsZero() {
		return 0
	}
	return ldt.time.Day()
}

// Hour 返回小时（0-23）
func (ldt LocalDateTime) Hour() int {
	if ldt.IsZero() {
		return 0
	}
	return ldt.time.Hour()
}

// Minute 返回分钟（0-59）
func (ldt LocalDateTime) Minute() int {
	if ldt.IsZero() {
		return 0
	}
	return ldt.time.Minute()
}

// Second 返回秒（0-59）
func (ldt LocalDateTime) Second() int {
	if ldt.IsZero() {
		return 0
	}
	return ldt.time.Second()
}

// Before 检查是否在另一个时间之前
func (ldt LocalDateTime) Before(other LocalDateTime) bool {
	if ldt.IsZero() || other.IsZero() {
		return false
	}
	return ldt.time.Before(*other.time)
}

// After 检查是否在另一个时间之后
func (ldt LocalDateTime) After(other LocalDateTime) bool {
	if ldt.IsZero() || other.IsZero() {
		return false
	}
	return ldt.time.After(*other.time)
}

// Equal 检查是否与另一个时间相等
func (ldt LocalDateTime) Equal(other LocalDateTime) bool {
	if ldt.IsZero() && other.IsZero() {
		return true
	}
	if ldt.IsZero() || other.IsZero() {
		return false
	}
	return ldt.time.Equal(*other.time)
}

// ToTime 转换为标准库 time.Time
func (ldt LocalDateTime) ToTime() (time.Time, bool) {
	if ldt.IsZero() {
		return zeroTime, false
	}
	return *ldt.time, true
}

// MarshalJSON 实现 json.Marshaler 接口
func (ldt LocalDateTime) MarshalJSON() ([]byte, error) {
	if ldt.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(ldt.time.Format("2006-01-02 15:04:05"))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ldt *LocalDateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "\"null\"" || s == "" {
		ldt.time = nil
		return nil
	}
	// 尝试带毫秒的格式
	t, err := time.Parse("2006-01-02 15:04:05.000000", s)
	if err == nil {
		ldt.time = &t
		return nil
	}
	// 尝试普通格式
	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		ldt.time = &t
		return nil
	}
	return fmt.Errorf("解析JSON日期时间失败: %v", err)
}

// Value 实现 driver.Valuer 接口，用于数据库操作
func (ldt LocalDateTime) Value() (driver.Value, error) {
	if ldt.IsZero() {
		return nil, nil
	}
	// 数据库序列化时使用6位毫秒数
	return ldt.time.Format("2006-01-02 15:04:05.000000"), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库操作
func (ldt *LocalDateTime) Scan(value any) error {
	if value == nil {
		ldt.time = nil
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		if v.IsZero() {
			ldt.time = nil
		} else {
			ldt.time = &v
		}
		return nil
	case string:
		if v == "0001-01-01 00:00:00" || v == "" {
			ldt.time = nil
			return nil
		}
		// 尝试带毫秒的格式
		t, err := time.Parse("2006-01-02 15:04:05.000000", v)
		if err == nil {
			ldt.time = &t
			return nil
		}
		// 尝试普通格式
		t, err = time.Parse("2006-01-02 15:04:05", v)
		if err == nil {
			ldt.time = &t
			return nil
		}
		return fmt.Errorf("解析数据库日期时间失败: %v", err)
	default:
		return errors.New("不支持的类型")
	}
}

// 辅助函数：将 Java 风格的格式符号转换为 Go 风格
func replaceJavaFormatSymbols(format string) string {
	replacements := map[string]string{
		"yyyy": "2006",
		"MM":   "01",
		"dd":   "02",
		"HH":   "15",
		"mm":   "04",
		"ss":   "05",
		"S":    "000", // 毫秒支持
	}
	for javaY, goY := range replacements {
		format = strings.ReplaceAll(format, javaY, goY)
	}
	return format
}
