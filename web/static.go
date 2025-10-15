package web

import (
	"github.com/dc-utils/args"
	"github.com/gofiber/fiber/v2"
	"time"
)

func staticMapping(app *fiber.App) {
	if args.GetBool("server.static.enabled", true) {
		staticPath := args.GetStr("server.static.path", "./static")
		compress := args.GetBool("server.static.compress", false)
		byteRange := args.GetBool("server.static.byte_range", false)
		browse := args.GetBool("server.static.browse", false)
		download := args.GetBool("server.static.download", false)
		index := args.GetStr("server.static.index", "")
		cacheDuration := args.GetDuration("server.static.cacheDuration", 10*time.Second)
		maxAge := args.GetInt("server.static.maxAge", 0)
		staticConfig := fiber.Static{
			Compress:      compress,
			ByteRange:     byteRange,
			Browse:        browse,
			Download:      download,
			Index:         index,
			CacheDuration: cacheDuration,
			MaxAge:        maxAge,
		}
		app.Static("/static", staticPath, staticConfig)
		app.Static("/", staticPath, staticConfig)
		app.Static("/favicon.ico", staticPath+"/favicon.ico", staticConfig)
	}
}
