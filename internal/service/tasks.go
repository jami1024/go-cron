package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-cron/config"
	"go-cron/internal/domain"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// TaskService 定义具备哪些方法
type TaskService interface {
	// Save 新增的返回为空的oldTask
	Save(ctx context.Context, task *domain.Task) (oldTask *domain.Task, err error)
}

type taskService struct {
	log     *zap.Logger
	taskMgr *domain.TaskMgr
}

func NewTaskService(l *zap.Logger, taskMgr *domain.TaskMgr) TaskService {
	return &taskService{
		log:     l,
		taskMgr: taskMgr,
	}
}

// Save 实现新增、编辑接口
func (t *taskService) Save(ctx context.Context, task *domain.Task) (oldTask *domain.Task, err error) {
	// 序列化
	taskJson, err := json.Marshal(task)
	if err != nil {
		t.log.Error(fmt.Sprintf("参数序列化失败,%v", err.Error()))
	}
	// etcd 中key的地址
	taskKey := config.Conf.EtcdConfig.KeyPath + task.Zk + "/" + task.Name
	//fmt.Println(string(taskJson))
	// 保存到etcd
	putResp, err := t.taskMgr.Kv.Put(context.TODO(), taskKey, string(taskJson), clientv3.WithPrevKV())
	if err != nil {
		fmt.Println("err", err.Error())
		return nil, err
	}
	// 如果是更新, 那么返回旧值
	if putResp.PrevKv != nil {
		// 对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldTask); err != nil {
			err = nil
			return
		}
	}
	return
}
