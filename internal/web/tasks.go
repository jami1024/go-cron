package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-cron/internal/domain"
	"go-cron/internal/service"
	"go.uber.org/zap"
)

// TaskHandler 我准备在它上面定义跟任务有关的路由
type TaskHandler struct {
	svc service.TaskService
	log *zap.Logger
}

func NewTaskHandler(svc service.TaskService, l *zap.Logger) *TaskHandler {

	return &TaskHandler{
		svc: svc,
		log: l,
	}
}

func (t *TaskHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/tasks")
	g.POST("/save", t.Save)

}

// Save 增加定时任务
// @Summary 增加定时任务
// @Description 增加定时任务
// @Accept  json
// @Produce  json
// @Param data body domain.Task true "请示参数data"
// @Success 200 {object} web.Result "请求成功"
// @Failure 400 {object} web.Result "请求错误"
// @Failure 500 {object} web.Result "内部错误"
// @Router /tasks/save [post]
func (t *TaskHandler) Save(ctx *gin.Context) {
	var req domain.Task

	// 获取请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		t.log.Error(fmt.Sprintln("参数异常", err))
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 100,
			Msg:  "参数异常",
			Data: err.Error(),
		})
		return
	}
	if req.Zk == "" {
		req.Zk = "zk"
	}

	t.log.Info(fmt.Sprintf("参数获取完毕,%v", req))

	oldTask, err := t.svc.Save(ctx, &req)
	if err != nil {
		printStr := fmt.Sprintf("任务添加错误,%s", err)
		t.log.Error(printStr)
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 200,
		Msg:  "添加完成",
		Data: oldTask,
	})
	return
}
