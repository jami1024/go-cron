package domain

import clientv3 "go.etcd.io/etcd/client/v3"

// Task 定时任务
type Task struct {
	Name     string `json:"name" binding:"required"`     //  任务名
	Command  string `json:"command" binding:"required"`  // shell命令
	CronExpr string `json:"cronExpr" binding:"required"` // cron表达式
	Zk       string `json:"zk"`                          //用来表示哪个中控，或者理解成不同时区的集群，其中cron表达式要根据zk来写
}

// TaskMgr 任务管理
type TaskMgr struct {
	Client *clientv3.Client `json:"client"`
	Kv     clientv3.KV      `json:"kv"`
	Lease  clientv3.Lease   `json:"lease"`
}
