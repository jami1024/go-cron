package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	// etcd删除，当目标不存在的时候也不会报错。
	g.POST("/delete", t.Delete)
	g.POST("/edit", t.Edit)
	g.GET("/list", t.List)

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

	req.UniqueCode = createUuid()

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

// Delete 删除定时任务
// @Summary 删除定时任务
// @Description 删除定时任务
// @Accept  json
// @Produce  json
// @Param data body domain.Task true "请示参数data"
// @Success 200 {object} web.Result "请求成功"
// @Failure 400 {object} web.Result "请求错误"
// @Failure 500 {object} web.Result "内部错误"
// @Router /tasks/delete [post]
func (t *TaskHandler) Delete(ctx *gin.Context) {
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
	if req.Zk == "" || req.UniqueCode == "" {
		t.log.Error(fmt.Sprintln("缺少参数"))
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 100,
			Msg:  "缺少参数",
			Data: "",
		})
		return
	}

	oldTask, err := t.svc.Delete(ctx, req.Name, req.Zk, req.UniqueCode)
	if err != nil {
		printStr := fmt.Sprintf("任务删除错误,%s", err)
		t.log.Error(printStr)
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 200,
		Msg:  "删除完成",
		Data: oldTask,
	})
	return
}

// Edit 编辑定时任务
// @Summary 编辑定时任务
// @Description 编辑定时任务
// @Accept  json
// @Produce  json
// @Param data body domain.Task true "请示参数data"
// @Success 200 {object} web.Result "请求成功"
// @Failure 400 {object} web.Result "请求错误"
// @Failure 500 {object} web.Result "内部错误"
// @Router /tasks/save [post]
func (t *TaskHandler) Edit(ctx *gin.Context) {
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
	if req.Zk == "" || req.UniqueCode == "" {
		t.log.Error(fmt.Sprintln("缺少参数"))
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 100,
			Msg:  "缺少参数",
			Data: "",
		})
		return
	}

	oldTask, err := t.svc.Save(ctx, &req)
	if err != nil {
		printStr := fmt.Sprintf("任务编辑错误,%s", err)
		t.log.Error(printStr)
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 200,
		Msg:  "修改完成",
		Data: oldTask,
	})
	return
}

// List 获取定时任务列表
// @Summary 获取定时任务列表
// @Description 获取定时任务列表
// @Accept  json
// @Produce  json
// @Param zk query string false "zk"
// @Success 200 {object} web.Result "请求成功"
// @Failure 400 {object} web.Result "请求错误"
// @Failure 500 {object} web.Result "内部错误"
// @Router /tasks/list [get]
func (t *TaskHandler) List(ctx *gin.Context) {
	var zk string
	// 获取请求参数
	zk = ctx.Query("zk")

	taskList, err := t.svc.List(ctx, zk)
	if err != nil {
		printStr := fmt.Sprintf("任务获取错误,%s", err)
		t.log.Error(printStr)
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 200,
		Msg:  "获取完毕",
		Data: taskList,
	})
	return
}

// createUuid 生成uuid
func createUuid() string {
	// 如需移除UUID中的连字符，可以使用如下代码
	uuidWithoutHyphen := uuid.New().String()
	uuidWithoutHyphen = strings.ReplaceAll(uuidWithoutHyphen, "-", "")
	return uuidWithoutHyphen
}
