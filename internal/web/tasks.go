package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-cron/internal/service"
)

// TaskHandler 我准备在它上面定义跟任务有关的路由
type TaskHandler struct {
	svc service.TaskService
}

func NewTaskHandler(svc service.TaskService) *TaskHandler {

	return &TaskHandler{
		svc: svc,
	}
}

func (t *TaskHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/tasks")
	g.POST("/save", t.Save)

}

func (t *TaskHandler) Save(ctx *gin.Context) {
	// 逻辑代码
	ctx.JSON(http.StatusOK, Result{
		Data: 1,
	})
}
