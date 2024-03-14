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
	Delete(ctx context.Context, taskName, zk, uniqueCode string) (oldTask *domain.Task, err error)
	List(ctx context.Context, zk string) (taskList []*domain.Task, err error)
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

// Save 实现任务新增、编辑接口
func (t *taskService) Save(ctx context.Context, task *domain.Task) (oldTask *domain.Task, err error) {
	// 序列化
	taskJson, err := json.Marshal(task)
	if err != nil {
		t.log.Error(fmt.Sprintf("参数序列化失败,%v", err.Error()))
	}
	// etcd 中key的地址
	taskKey := config.Conf.EtcdConfig.KeyPath + task.Zk + "/" + task.Name + "_" + task.UniqueCode
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

// Delete 实现任务删除接口
func (t *taskService) Delete(ctx context.Context, taskName, zk, uniqueCode string) (oldTask *domain.Task, err error) {

	// etcd 中key的地址
	taskKey := config.Conf.EtcdConfig.KeyPath + zk + "/" + taskName + "_" + uniqueCode
	fmt.Println("删除任务地址", taskKey)
	// 删除
	delResp, err := t.taskMgr.Kv.Delete(context.TODO(), taskKey, clientv3.WithPrevKV())
	if err != nil {
		fmt.Println("任务删除失败", err.Error())
		t.log.Error(fmt.Sprintf("任务删除失败,%s", err.Error()))
		return nil, err
	}

	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		// 解析一下旧值, 返回它
		err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldTask)
		if err != nil {
			fmt.Println("任务删除失败", err.Error())
			t.log.Error(fmt.Sprintf("任务删除失败,%s", err.Error()))
			return nil, err
		}
	}
	return
}

// List 实现获取任务接口
func (t *taskService) List(ctx context.Context, zk string) (taskList []*domain.Task, err error) {

	// 按zk搜索
	var dirKey string
	if zk == "" {
		// 任务保存的目录
		dirKey = config.Conf.EtcdConfig.KeyPath
	} else {
		dirKey = config.Conf.EtcdConfig.KeyPath + zk + "/"
	}
	// 获取目录下所有任务信息
	getResp, err := t.taskMgr.Kv.Get(context.TODO(), dirKey, clientv3.WithPrefix())
	if err != nil {
		fmt.Println("任务获取失败", err.Error())
		t.log.Error(fmt.Sprintf("任务获取失败,%s", err.Error()))
		return nil, err
	}
	// 初始化数组空间
	taskList = make([]*domain.Task, 0)
	// 遍历所有任务, 进行反序列化
	for _, kvPair := range getResp.Kvs {
		task := &domain.Task{}
		err = json.Unmarshal(kvPair.Value, task)
		// 解析错误不报异常
		if err != nil {
			err = nil
			continue
		}
		taskList = append(taskList, task)
	}
	return
}
