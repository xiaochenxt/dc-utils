package ptr

import (
	"reflect"
	"time"
)

// String 返回字符串指针
func String(s string) *string {
	return &s
}

// Int 返回整数指针
func Int(i int) *int {
	return &i
}

// Int8 返回 int8 指针
func Int8(i int8) *int8 {
	return &i
}

// Int16 返回 int16 指针
func Int16(i int16) *int16 {
	return &i
}

// Int32 返回 int32 指针
func Int32(i int32) *int32 {
	return &i
}

// Int64 返回 int64 指针
func Int64(i int64) *int64 {
	return &i
}

// Uint 返回 uint 指针
func Uint(i uint) *uint {
	return &i
}

// Uint8 返回 uint8 指针
func Uint8(i uint8) *uint8 {
	return &i
}

// Uint16 返回 uint16 指针
func Uint16(i uint16) *uint16 {
	return &i
}

// Uint32 返回 uint32 指针
func Uint32(i uint32) *uint32 {
	return &i
}

// Uint64 返回 uint64 指针
func Uint64(i uint64) *uint64 {
	return &i
}

// Float32 返回 float32 指针
func Float32(f float32) *float32 {
	return &f
}

// Float64 返回 float64 指针
func Float64(f float64) *float64 {
	return &f
}

// Bool 返回布尔指针
func Bool(b bool) *bool {
	return &b
}

// Time 返回时间指针
func Time(t time.Time) *time.Time {
	return &t
}

// Map 返回Map指针
func Map(m map[string]any) *map[string]any {
	return &m
}

// StringValue 安全获取字符串指针的值，指针为 nil 时返回默认值
func StringValue(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// IntValue 安全获取整数指针的值，指针为 nil 时返回默认值
func IntValue(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// Int64Value 安全获取 int64 指针的值，指针为 nil 时返回默认值
func Int64Value(ptr *int64, defaultValue int64) int64 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// BoolValue 安全获取布尔指针的值，指针为 nil 时返回默认值
func BoolValue(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// TimeValue 安全获取时间指针的值，指针为 nil 时返回默认值
func TimeValue(ptr *time.Time, defaultValue time.Time) time.Time {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// Value 获取指针的值，明确非nil时使用，空指针时会报错
func Value[T any](ptr *T) T {
	return *ptr
}

// IsNil 检查指针是否为 nil（支持所有指针类型）
func IsNil(ptr any) bool {
	if ptr == nil {
		return true
	}
	value := reflect.ValueOf(ptr)
	return value.Kind() == reflect.Ptr && value.IsNil()
}
