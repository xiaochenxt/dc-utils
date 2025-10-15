package ip

import (
	"encoding/binary"
	"github.com/gofiber/fiber/v2"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"net"
	"strings"
)

func GetRemoteIp(c *fiber.Ctx) string {
	var ip = c.GetReqHeaders()["x-real-ip"]
	if ip == nil || len(ip) == 0 {
		return c.IP()
	}
	return ip[0]
}

// IPv4ToUint64 将 IPv4 地址字符串转换为 uint64 数值
// 示例："192.168.1.1" → 3232235777
func IPv4ToUint64(ip string) uint64 {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0 // 无效 IP
	}

	// 转换为 IPv4 格式（去除 IPv6 前缀）
	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0 // 不是有效的 IPv4 地址
	}

	// 将 []byte 转换为 uint32，再提升为 uint64
	return uint64(binary.BigEndian.Uint32(ipv4))
}

// Uint64ToIPv4 将 uint64 数值转换为 IPv4 地址字符串
// 示例：3232235777 → "192.168.1.1"
func Uint64ToIPv4(ipNum uint64) string {
	// 确保数值在 IPv4 范围内（0~4294967295）
	if ipNum > 0xFFFFFFFF {
		return "" // 超出 IPv4 范围
	}

	// 将 uint32 转换为 []byte
	ipBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ipBytes, uint32(ipNum))

	// 转换为 IP 字符串
	return net.IP(ipBytes).String()
}

var searcher *xdb.Searcher

func init() {
	var cBuff, _ = xdb.LoadContentFromFile("china.xdb")
	searcher, _ = xdb.NewWithBuffer(cBuff)
}

type IpInfo struct {
	Ip                    string `json:"ip"`
	Continent             string `json:"continent"`
	Country               string `json:"country"`
	Province              string `json:"province"`
	City                  string `json:"city"`
	District              string `json:"district"`
	Isp                   string `json:"isp"`
	ZoningCode1           string `json:"zoning_code_1"`
	ZoningCode2           string `json:"zoning_code_2"`
	ZoningCode3           string `json:"zoning_code_3"`
	NationalEnglish       string `json:"national_english"`
	CountryAbbreviations  string `json:"country_abbreviations"`
	InternationalAreaCode string `json:"international_area_code"`
	Longitude             string `json:"longitude"`
	Latitude              string `json:"latitude"`
}

func GetInfo(ip string) *IpInfo {
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	ipInfoStr, _ := searcher.SearchByStr(ip)
	if ipInfoStr == "" {
		return &IpInfo{
			Ip: ip,
		}
	}
	ipInfoArr := strings.Split(ipInfoStr, "|")
	ipInfo := &IpInfo{
		Ip:                    ip,
		Continent:             ipInfoArr[0],
		Country:               ipInfoArr[1],
		Province:              ipInfoArr[2],
		City:                  ipInfoArr[3],
		District:              ipInfoArr[4],
		Isp:                   ipInfoArr[5],
		ZoningCode1:           ipInfoArr[6],
		ZoningCode2:           ipInfoArr[7],
		ZoningCode3:           ipInfoArr[8],
		NationalEnglish:       ipInfoArr[9],
		CountryAbbreviations:  ipInfoArr[10],
		InternationalAreaCode: ipInfoArr[11],
		Longitude:             ipInfoArr[12],
		Latitude:              ipInfoArr[13],
	}
	return ipInfo
}
