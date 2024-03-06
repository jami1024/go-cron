package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"

	ginSwagger "github.com/swaggo/gin-swagger"
	"go-cron/config"
	_ "go-cron/docs"
	"go-cron/pkg/logger"
)

func Setup(mode string) *gin.Engine {
	// gin 模式
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	//r.GET("/", func(c *gin.Context) {
	//	c.String(http.StatusOK, "hello,gin!!!")
	//})
	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Conf.Version)
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
