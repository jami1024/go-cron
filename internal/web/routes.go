package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-cron/config"
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
	return r
}
