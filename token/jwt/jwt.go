package jwt

import (
	"encoding/json"
	"github.com/dc-utils/args"
	"github.com/dc-utils/snowflake"
	"github.com/dc-utils/token"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var tokenName string

var secret string

var enabledRefresh = true

var expireDuration time.Duration

func init() {
	tokenName = args.GetStr("token.name", "dc-token")
	enabledRefresh = args.GetBool("token.refresh.enabled", true)
	expireDuration = args.GetDuration("token.expire.seconds", 7200*time.Second)
	secret = args.Get("token.secret")
	if secret == "" {
		secret = uuid.New().String()
		log.Warnf("jwt密钥为空，随机生成一个 %s，生产环境使用必需配置", secret)
	}
}

type TokenInfo struct {
	Token                 string  `json:"token"`
	ExpiresIn             int64   `json:"expires_in"`
	RefreshToken          *string `json:"refresh_token"`
	RefreshTokenExpiresIn *int64  `json:"refresh_token_expires_in"`
}

func GetToken(c *fiber.Ctx) string {
	tk := c.Get(tokenName)
	if tk != "" {
		return tk
	}
	tk = c.Cookies(tokenName)
	if tk != "" {
		return tk
	}
	tk = c.Get("Authorization")
	if tk != "" {
		return tk[7:]
	}
	return c.Query(tokenName)
}

func CreateToken(authentication token.Principal, expireDuration time.Duration) *TokenInfo {
	expiresIn := time.Now().Add(expireDuration).UnixMilli()
	claims := jwt.MapClaims{
		"exp":     expiresIn,
		"payload": authentication,
	}
	jwtInfo := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tk, _ := jwtInfo.SignedString([]byte(secret))
	refreshTokenExpiresIn := expiresIn + 3600*24*1000
	refreshToken := createRefreshToken(authentication, refreshTokenExpiresIn)
	return &TokenInfo{
		Token:                 tk,
		ExpiresIn:             expiresIn,
		RefreshToken:          &refreshToken,
		RefreshTokenExpiresIn: &refreshTokenExpiresIn,
	}
}

func RefreshToken(refreshToken string) *TokenInfo {
	if !enabledRefresh {
		log.Error("未开启刷新token功能")
		return nil
	}
	parse, _ := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	jwtInfo := parse.Claims.(jwt.MapClaims)
	if jwtInfo["ait"] == nil {
		log.Error("不是一个刷新token，" + refreshToken)
		return nil
	}
	return CreateToken(jwtInfo["payload"].(token.Principal), expireDuration)
}

func createRefreshToken(authentication token.Principal, exp int64) string {
	jwtInfo := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.Claims(jwt.MapClaims{
		"exp":     exp,
		"ati":     snowflake.NextId(),
		"payload": authentication,
	}))
	tk, _ := jwtInfo.SignedString([]byte(secret))
	return tk
}

func ReadToken(tk string, principal token.Principal) token.Principal {
	parse, _ := jwt.Parse(tk, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	jwtInfo := parse.Claims.(jwt.MapClaims)
	payload := jwtInfo["payload"].(map[string]any)
	marshal, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(marshal, &principal)
	if err != nil {
		return nil
	}
	return principal
}
