package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"go-cron/internal/service"
	"go.uber.org/zap"

	ginSwagger "github.com/swaggo/gin-swagger"
	"go-cron/config"
	_ "go-cron/docs"
	"go-cron/pkg/logger"
)

func InitWeb(zapL *zap.Logger) *gin.Engine {
	// gin 模式

	if config.Conf.Mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	server := gin.New()
	server.Use(logger.GinLogger(), logger.GinRecovery(true))

	server.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Conf.Version)
	})
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 任务路由
	//userHandler := web.NewUserHandler(userService, codeService, handler)
	// service.NewUserService(userRepository, loggerV1)
	taskHandler := NewTaskHandler(service.NewTaskService(zapL))
	taskHandler.RegisterRoutes(server)
	return server
}
