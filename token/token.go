package token

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dc-utils/args"
	"github.com/dc-utils/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
)

func init() {
	tokenName = args.GetStr("token.name", "dc-token")
	enabledRefresh = args.GetBool("token.refresh.enabled", true)
	expireDuration = args.GetDuration("token.expire.seconds", 7200*time.Second)
}

var tokenName string

var enabledRefresh bool

var expireDuration time.Duration

var tokenPrefix = "dc:security:token:"

var tokenInfoPrefix = "dc:security:token:info:"

var onlineScript = fmt.Sprintf(`
-- Token
            local token = ARGV[1]
            
            -- 用户唯一标识
            local userId = ARGV[2]
            
            -- Token信息
            local payload = ARGV[3]
            
            -- 最大在线数限制作为参数传入
            local maxOnlineNum = tonumber(ARGV[4])
            
            -- token有效期（单位秒）
            local expireTimeInSec = tonumber(ARGV[5])
            
            -- Token的键前缀
            local TOKEN_PREFIX = "%s"
            -- Token信息的键前缀
            local TOKEN_INFO_PREFIX = "%s"
            
            -- 标记在线并设置过期时间
            redis.call("SET", TOKEN_PREFIX .. token, "%s", "EX", expireTimeInSec)
            
            -- 保存Token信息，并设置过期时间
            redis.call("SET", TOKEN_INFO_PREFIX .. userId, payload, "EX", expireTimeInSec);
            
            -- 使用List保存Token列表
            local onlineListKey = TOKEN_PREFIX .. userId
            -- 将token写入Token列表
            redis.call("RPUSH", onlineListKey, token)
            -- 设置Token列表的过期时间
            redis.call("EXPIRE", onlineListKey, expireTimeInSec)
            
            -- 获取在线数量
            local onlineNum = redis.call("LLEN", onlineListKey)

            -- 检查并限制在线数量
            if onlineNum > maxOnlineNum then
                local onlineToken = redis.call("LPOP", onlineListKey)
                redis.call("UNLINK", TOKEN_PREFIX .. onlineToken)
            end
`, tokenPrefix, tokenInfoPrefix, "1")

var readTokenScript = fmt.Sprintf(`
local TOKEN_PREFIX = "%s"
local TOKEN_INFO_PREFIX = "%s"
-- Token
local token = ARGV[1]
-- 用户唯一标识
local userId = ARGV[2]
local tokenKey = TOKEN_PREFIX .. token
local tokenValue = redis.call("GET", tokenKey)
if tokenValue then
-- 如果是在线状态，则返回Token信息
if tokenValue == "%s" then
return redis.call("GET", TOKEN_INFO_PREFIX .. userId)
else
-- 不是在线状态就删除Token并返回当前状态值
redis.call("UNLINK", tokenKey)
return tokenValue
end
end
return nil
`, tokenPrefix, tokenInfoPrefix, "1")

var offlineScript = fmt.Sprintf(`
local TOKEN_PREFIX = "%s"
                    local TOKEN_INFO_PREFIX = "%s"
                    -- 用户唯一标识
                    local userId = ARGV[1]
                    local onlineListKey = TOKEN_PREFIX .. userId
                    local tokenList = redis.call("LRANGE", onlineListKey, 0, -1)
                    for i, key in ipairs(tokenList) do
                        redis.call("SET", TOKEN_PREFIX .. key, "%s", "EX", 30)
                    end
                    redis.call("UNLINK", onlineListKey)
                    redis.call("UNLINK", TOKEN_INFO_PREFIX .. userId)
`, tokenPrefix, tokenInfoPrefix, "0")

var renewScript = fmt.Sprintf(`-- Token
                    local token = ARGV[1]
                    local TOKEN_PREFIX = "%s"
                    local tokenKey = TOKEN_PREFIX .. token;
                    local ttl = redis.call("TTL", tokenKey)
                    if ttl > 0 and ttl < 600 then
                        -- 用户唯一标识
                        local userId = ARGV[2]
                        -- 过期时间
                        local expireTimeInSec = tonumber(ARGV[3])
                        local TOKEN_INFO_PREFIX = "%s"
                        redis.call("EXPIRE", TOKEN_INFO_PREFIX .. userId, expireTimeInSec)
                        redis.call("EXPIRE", TOKEN_PREFIX .. userId, expireTimeInSec)
                        local res = tonumber(redis.call("EXPIRE", tokenKey, expireTimeInSec))
                        if res > 0 then
                            return expireTimeInSec
                        end
                    end
                    return ttl`, tokenPrefix, tokenInfoPrefix)

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
	return c.Query(tokenName)
}

func CreateToken(authentication Principal) *TokenInfo {
	expiresIn := time.Now().Add(expireDuration).UnixMilli()
	jsonBytes, err := json.Marshal(authentication)
	if err != nil {
		return nil
	}
	u, _ := uuid.NewV7()
	tk := authentication.GetName() + "-" + strings.ReplaceAll(u.String(), "-", "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	redis.Get().Eval(ctx, onlineScript, []string{}, []string{tk, authentication.GetName(), string(jsonBytes), "5", strconv.FormatInt(int64(expireDuration.Seconds()), 10)})
	return &TokenInfo{
		Token:     tk,
		ExpiresIn: expiresIn,
	}
}

func RefreshToken(tk string) *TokenInfo {
	if !enabledRefresh {
		log.Error("未开启刷新token功能")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := redis.Get().Eval(ctx, renewScript, []string{}, []string{tk, getName(tk), strconv.FormatInt(int64(expireDuration.Seconds()), 10)}).Int64()
	if err != nil {
		return nil
	}
	return &TokenInfo{
		Token:     tk,
		ExpiresIn: time.Now().Add(time.Duration(res) * time.Second).UnixMilli(),
	}
}

func ReadToken(tk string, principal Principal) Principal {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tokenInfo := redis.Get().Eval(ctx, readTokenScript, []string{}, []string{tk, getName(tk)}).String()
	if tokenInfo == "" {
		return nil
	} else if tokenInfo == "0" {
		return nil
	}
	err := json.Unmarshal([]byte(tokenInfo), principal)
	if err != nil {
		return nil
	}
	return principal
}

func Offline(userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	redis.Get().Eval(ctx, offlineScript, []string{}, []string{userId})
}

func RefreshAuthentication(authentication Principal) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	jsonBytes, err := json.Marshal(authentication)
	if err != nil {
		return
	}
	redis.Get().Set(ctx, tokenInfoPrefix+authentication.GetName(), jsonBytes, 0)
}

func getName(tk string) string {
	if tk == "" {
		return ""
	}
	return strings.Split(tk, "-")[0]
}
