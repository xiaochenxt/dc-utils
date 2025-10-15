package converter

import (
	"database/sql"
	"encoding/json"
	"math"
	"strconv"
	"time"
)

// ToString 将任意类型转换为字符串
func ToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format(time.RFC3339)
	case nil:
		return ""
	default:
		jsonStr, err := AnyToJSON(v)
		if err != nil {
			return ""
		}
		return jsonStr
	}
}

// ToInt 将值转换为 int 类型
func ToInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		if v < math.MinInt || v > math.MaxInt {
			return 0, strconv.ErrRange
		}
		return int(v), nil
	case uint:
		if v > uint(math.MaxInt) {
			return 0, strconv.ErrRange
		}
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		if v > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return int(v), nil
	case uint64:
		if v > uint64(math.MaxInt) {
			return 0, strconv.ErrRange
		}
		return int(v), nil
	case float32:
		if v < float32(math.MinInt) || v > float32(math.MaxInt) {
			return 0, strconv.ErrRange
		}
		return int(v), nil
	case float64:
		if v < float64(math.MinInt) || v > float64(math.MaxInt) {
			return 0, strconv.ErrRange
		}
		return int(v), nil
	case string:
		i64, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return 0, err
		}
		if i64 < math.MinInt || i64 > math.MaxInt {
			return 0, strconv.ErrRange
		}
		return int(i64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToInt8 将值转换为 int8 类型
func ToInt8(value any) (int8, error) {
	switch v := value.(type) {
	case int:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case int8:
		return v, nil
	case int16:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case int32:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case int64:
		if v < math.MinInt8 || v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case uint:
		if v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case uint8:
		if v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case uint16:
		if v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case uint32:
		if v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case uint64:
		if v > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case float32:
		if v < float32(math.MinInt8) || v > float32(math.MaxInt8) {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case float64:
		if v < float64(math.MinInt8) || v > float64(math.MaxInt8) {
			return 0, strconv.ErrRange
		}
		return int8(v), nil
	case string:
		i64, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return 0, err
		}
		return int8(i64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToInt16 将值转换为 int16 类型
func ToInt16(value any) (int16, error) {
	switch v := value.(type) {
	case int:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case int8:
		return int16(v), nil
	case int16:
		return v, nil
	case int32:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case int64:
		if v < math.MinInt16 || v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case uint:
		if v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case uint8:
		return int16(v), nil
	case uint16:
		if v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case uint32:
		if v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case uint64:
		if v > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case float32:
		if v < float32(math.MinInt16) || v > float32(math.MaxInt16) {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case float64:
		if v < float64(math.MinInt16) || v > float64(math.MaxInt16) {
			return 0, strconv.ErrRange
		}
		return int16(v), nil
	case string:
		i64, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return 0, err
		}
		return int16(i64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToInt32 将值转换为 int32 类型
func ToInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case uint:
		// 修正：先将 uint 转换为 uint64 再比较
		if uint64(v) > uint64(math.MaxInt32) {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		// 修正：使用 math.MaxInt32 而不是 math.MaxInt
		if v > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case uint64:
		if v > uint64(math.MaxInt32) {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case float32:
		if v < float32(math.MinInt32) || v > float32(math.MaxInt32) {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case float64:
		if v < float64(math.MinInt32) || v > float64(math.MaxInt32) {
			return 0, strconv.ErrRange
		}
		return int32(v), nil
	case string:
		i64, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(i64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToInt64 将值转换为 int64 类型
func ToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, strconv.ErrRange
		}
		return int64(v), nil
	case float32:
		if v < float32(math.MinInt64) || v > float32(math.MaxInt64) {
			return 0, strconv.ErrRange
		}
		return int64(v), nil
	case float64:
		if v < float64(math.MinInt64) || v > float64(math.MaxInt64) {
			return 0, strconv.ErrRange
		}
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToFloat64 将值转换为 float64 类型
func ToFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToBool 将值转换为 bool 类型
func ToBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, strconv.ErrSyntax
	}
}

// ToUint 将值转换为 uint 类型
func ToUint(value any) (uint, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case int8:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case int16:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case int32:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case int64:
		if v < 0 || v > math.MaxInt64 {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		if v > uint64(math.MaxUint) {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case float32:
		if v < 0 || v > float32(math.MaxUint) {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case float64:
		if v < 0 || v > float64(math.MaxUint) {
			return 0, strconv.ErrRange
		}
		return uint(v), nil
	case string:
		u64, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return 0, err
		}
		if u64 > uint64(math.MaxUint) {
			return 0, strconv.ErrRange
		}
		return uint(u64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToUint8 将值转换为 uint8 类型
func ToUint8(value any) (uint8, error) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case int8:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case int16:
		if v < 0 || v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case int32:
		if v < 0 || v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case int64:
		if v < 0 || v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case uint:
		if v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case uint8:
		return v, nil
	case uint16:
		if v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case uint32:
		if v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case uint64:
		if v > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case float32:
		if v < 0 || v > float32(math.MaxUint8) {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case float64:
		if v < 0 || v > float64(math.MaxUint8) {
			return 0, strconv.ErrRange
		}
		return uint8(v), nil
	case string:
		u64, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(u64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

func ToByte(value any) (byte, error) {
	return ToUint8(value)
}

// ToUint16 将值转换为 uint16 类型
func ToUint16(value any) (uint16, error) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case int8:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case int16:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case int32:
		if v < 0 || v > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case int64:
		if v < 0 || v > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case uint:
		if v > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case uint8:
		return uint16(v), nil
	case uint16:
		return v, nil
	case uint32:
		if v > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case uint64:
		if v > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case float32:
		if v < 0 || v > float32(math.MaxUint16) {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case float64:
		if v < 0 || v > float64(math.MaxUint16) {
			return 0, strconv.ErrRange
		}
		return uint16(v), nil
	case string:
		u64, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(u64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToUint32 将值转换为 uint32 类型
func ToUint32(value any) (uint32, error) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case int8:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case int16:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case int32:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case int64:
		if v < 0 || v > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case uint:
		// 修正：先将 uint 转换为 uint64 再比较
		if uint64(v) > uint64(math.MaxUint32) {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case uint16:
		return uint32(v), nil
	case uint32:
		return v, nil
	case uint64:
		if v > math.MaxUint32 {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case float32:
		if v < 0 || v > float32(math.MaxUint32) {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case float64:
		if v < 0 || v > float64(math.MaxUint32) {
			return 0, strconv.ErrRange
		}
		return uint32(v), nil
	case string:
		u64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(u64), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// ToUint64 将值转换为 uint64 类型
func ToUint64(value any) (uint64, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case int8:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case int16:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		if v < 0 || v > float32(math.MaxUint64) {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case float64:
		if v < 0 || v > float64(math.MaxUint64) {
			return 0, strconv.ErrRange
		}
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

// JSONToMap 将 JSON 字符串转换为 map[string]any
func JSONToMap(jsonStr string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// JSONToAny 将 JSON 字符串转换为任意类型
func JSONToAny(jsonStr string) (any, error) {
	var result any
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// JSONToBytes 将 JSON 字符串转换为字节切片
func JSONToBytes(jsonStr string) ([]byte, error) {
	return []byte(jsonStr), nil
}

// MapToJSON 将 map 转换为 JSON 字符串
func MapToJSON(m map[string]any) (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// AnyToJSON 将任意类型转换为 JSON 字符串
func AnyToJSON(v any) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// BytesToJSON 将字节切片转换为 JSON 字符串
func BytesToJSON(b []byte) (string, error) {
	var result any
	err := json.Unmarshal(b, &result)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// NullString 将字符串转换为sql.NullString
func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// NullInt64 将int64转换为sql.NullInt64
func NullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}

// NullFloat64 将float64转换为sql.NullFloat64
func NullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   true,
	}
}

// NullBool 将bool转换为sql.NullBool
func NullBool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: true,
	}
}

// NullTime 将time.Time转换为sql.NullTime
func NullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

// NullByteSlice 将字节切片转换为sql.RawBytes
func NullByteSlice(b []byte) sql.RawBytes {
	if b == nil {
		return nil
	}
	return append([]byte(nil), b...) // 复制数据
}
