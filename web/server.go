package web

import (
	"context"
	"github.com/dc-utils/args"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Start 启动Web服务端
//
// callback: 在回调函数中完成注册路由，中间件等操作
func Start(callback func(app *fiber.App)) {
	StartWithConfig(fiber.Config{
		ReadTimeout:                  args.GetDuration("server.readTimeout", 5*time.Second),
		WriteTimeout:                 args.GetDuration("server.writeTimeout", 0),
		IdleTimeout:                  args.GetDuration("server.idleTimeout", 0),
		ReadBufferSize:               args.GetInt("server.readBufferSize", 0),
		WriteBufferSize:              args.GetInt("server.writeBufferSize", 0),
		BodyLimit:                    args.GetInt("server.bodyLimit", 0),
		Concurrency:                  args.GetInt("server.concurrency", 0),
		Prefork:                      args.GetBool("server.prefork", false),
		DisableStartupMessage:        args.GetBool("server.disableStartupMessage", false),
		EnablePrintRoutes:            args.GetBool("server.enablePrintRoutes", false),
		GETOnly:                      args.GetBool("server.getOnly", false),
		Immutable:                    args.GetBool("server.immutable", false),
		PassLocalsToViews:            args.GetBool("server.passLocalsToViews", false),
		ProxyHeader:                  args.GetStr("server.proxyHeader", ""),
		ServerHeader:                 args.GetStr("server.serverHeader", ""),
		StrictRouting:                args.GetBool("server.strictRouting", false),
		UnescapePath:                 args.GetBool("server.unescapePath", false),
		DisablePreParseMultipartForm: args.GetBool("server.disablePreParseMultipartForm", false),
		AppName:                      args.GetStr("server.name", ""),
		CompressedFileSuffix:         args.GetStr("server.compressedFileSuffix", ""),
	}, callback)
}

// StartWithConfig 配置参数后并启动Web服务端（当需要配置一些依赖其他包的参数时使用，如第三方json、xml）
//
// callback: 在回调函数中完成注册路由，中间件等操作
func StartWithConfig(config fiber.Config, callback func(app *fiber.App)) {
	webPort := args.GetStr("server.port", "8080")
	app := fiber.New(config)
	if args.GetBool("server.recover.enabled", true) {
		app.Use(recover.New())
	}
	if args.GetBool("server.cors.enabled", true) {
		app.Use(cors.New())
	}
	if args.GetBool("server.compress.enabled", false) {
		app.Use(compress.New())
	}
	if args.GetBool("server.etag.enabled", false) {
		app.Use(etag.New())
	}
	callback(app)
	staticMapping(app)
	go func() {
		if err := app.Listen(":" + webPort); err != nil {
			log.Fatalf("web服务运行失败: %v\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), args.GetDuration("server.shutdown.timeout", 30*time.Second))
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("web服务正常停止异常:%v", err)
	}
	log.Info("web服务已正常停止")
}
