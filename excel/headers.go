package excel

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// SetXLSXHeaders 设置XLSX文件下载的HTTP响应头
func SetXLSXHeaders(c *fiber.Ctx, fileName string) {
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
}

// SetCSVHeaders 设置CSV文件下载的HTTP响应头
func SetCSVHeaders(c *fiber.Ctx, fileName string) {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
}
