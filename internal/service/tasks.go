package service

import (
	"context"
	"fmt"

	"go-cron/internal/domain"
	"go.uber.org/zap"
)

// TaskService 定义具备哪些方法
type TaskService interface {
	Save(ctx context.Context, email, password string) (domain.Task, error)
}

type taskService struct {
	log *zap.Logger
}

func NewTaskService(l *zap.Logger) *taskService {
	return &taskService{
		log: l,
	}
}

// Save 实现新增、编辑接口
func (t *taskService) Save(ctx context.Context, email, password string) (domain.Task, error) {
	//TODO implement me etcd操作
	fmt.Println("xxxx")
	panic("implement me")
}
