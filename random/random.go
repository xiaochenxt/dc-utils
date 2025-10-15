package random

import (
	"math/rand"
)

// 预定义字符集
const (
	Digits       = "0123456789"
	Lowercase    = "abcdefghijklmnopqrstuvwxyz"
	Uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	SpecialChars = "!@#$%^&*()-_=+,.?/:;{}[]`~"
)

// String 生成指定长度的随机字符串，可指定字符集
func String(length int, charset string) string {
	if length <= 0 {
		return ""
	}
	if charset == "" {
		charset = Digits + Lowercase + Uppercase
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// StringDigits 生成纯数字随机字符串
func StringDigits(length int) string {
	return String(length, Digits)
}

// StringLowercase 生成纯小写字母随机字符串
func StringLowercase(length int) string {
	return String(length, Lowercase)
}

// StringUppercase 生成纯大写字母随机字符串
func StringUppercase(length int) string {
	return String(length, Uppercase)
}

// StringAlphaNumeric 生成字母数字混合随机字符串
func StringAlphaNumeric(length int) string {
	return String(length, Digits+Lowercase+Uppercase)
}

// StringWithSpecial 生成包含特殊字符的随机字符串
func StringWithSpecial(length int) string {
	return String(length, Digits+Lowercase+Uppercase+SpecialChars)
}

// Int 生成指定范围内的随机整数 [min, max)
func Int(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min)
}

// Int64 生成指定范围内的随机64位整数 [min, max)
func Int64(min, max int64) int64 {
	if min >= max {
		return min
	}
	return min + rand.Int63n(max-min)
}

// IntSlice 生成指定数量、指定范围内的随机整数切片
func IntSlice(count, min, max int) []int {
	s := make([]int, count)
	for i := range s {
		s[i] = Int(min, max)
	}
	return s
}

// Shuffle 随机打乱字符串顺序
func Shuffle(s string) string {
	runes := []rune(s)
	rand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})
	return string(runes)
}

// Password 生成强密码（包含数字、大小写字母和特殊字符）
func Password(length int) string {
	if length < 4 {
		return StringWithSpecial(length)
	}

	// 确保至少包含一个数字、小写字母、大写字母和特殊字符
	password := make([]rune, length)
	password[0] = rune(Digits[rand.Intn(len(Digits))])
	password[1] = rune(Lowercase[rand.Intn(len(Lowercase))])
	password[2] = rune(Uppercase[rand.Intn(len(Uppercase))])
	password[3] = rune(SpecialChars[rand.Intn(len(SpecialChars))])

	// 填充剩余字符
	charset := Digits + Lowercase + Uppercase + SpecialChars
	for i := 4; i < length; i++ {
		password[i] = rune(charset[rand.Intn(len(charset))])
	}

	// 随机打乱顺序
	return Shuffle(string(password))
}
