package domain

import (
	"context"
	"time"

	"github.com/gorhill/cronexpr"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Task 定时任务
type Task struct {
	Name       string `json:"name" binding:"required"`     //  任务名
	Command    string `json:"command" binding:"required"`  // shell命令
	CronExpr   string `json:"cronExpr" binding:"required"` // cron表达式
	Zk         string `json:"zk"`                          //用来表示哪个中控，或者理解成不同时区的集群，其中cron表达式要根据zk来写
	UniqueCode string `json:"uniqueCode"`                  // 用来拼接任务名称保证任务唯一，用户不用关注、程序自身赋值。
}

// TaskMgr 任务管理
type TaskMgr struct {
	Client *clientv3.Client `json:"client"`
	Kv     clientv3.KV      `json:"kv"`
	Lease  clientv3.Lease   `json:"lease"`
}

// WorkerTaskMgr worker节点任务管理器
type WorkerTaskMgr struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

// TaskEvent 任务变化事件
type TaskEvent struct {
	EventType int //  SAVE, DELETE
	Task      *Task
}

// BuildTaskExecuteInfo 构造执行状态信息
func BuildTaskExecuteInfo(taskSchedulePlan *TaskSchedulePlan) (taskExecuteInfo *TaskExecuteInfo) {
	taskExecuteInfo = &TaskExecuteInfo{
		Task:     taskSchedulePlan.Task,
		PlanTime: taskSchedulePlan.NextTime, // 计算调度时间
		RealTime: time.Now(),                // 真实调度时间
	}
	taskExecuteInfo.CancelCtx, taskExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

// TaskSchedulePlan 任务调度计划
type TaskSchedulePlan struct {
	Task     *Task                // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cronexpr表达式
	NextTime time.Time            // 下次调度时间
}

// TaskExecuteInfo 任务执行状态信息
type TaskExecuteInfo struct {
	Task       *Task              // 任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 任务command的context
	CancelFunc context.CancelFunc //  用于取消command执行的cancel函数
}

// TaskExecuteResult 任务执行结果
type TaskExecuteResult struct {
	ExecuteInfo *TaskExecuteInfo // 执行状态
	Output      []byte           // 脚本输出
	Err         error            // 脚本错误原因
	StartTime   time.Time        // 启动时间
	EndTime     time.Time        // 结束时间
}
