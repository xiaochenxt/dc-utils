package bcrypt

import (
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/crypto/bcrypt"
)

func GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, cost)
}

func CompareHashAndPassword(hash, password []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func Encode(password string, cost int) string {
	hash, err := GenerateFromPassword([]byte(password), cost)
	if err != nil {
		// 异常罕见，不用考虑
		log.Errorf("密码生成哈希值失败，返回空字符串，%v", err)
		return ""
	}
	return string(hash)
}

func Encode4(password string) string {
	return Encode(password, 4)
}

func Encode10(password string) string {
	return Encode(password, 10)
}

// Matches 密码匹配，成功返回true，失败返回false
//
// encodedPassword：哈希过后的密码hash值
//
// password 原始密码
func Matches(encodedPassword, password string) bool {
	res, err := CompareHashAndPassword([]byte(encodedPassword), []byte(password))
	if err != nil {
		log.Debugf("密码匹配失败，%v", err)
		return false
	}
	return res
}
