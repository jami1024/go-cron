package worker

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// TaskLock 分布式锁(TXN事务)
type TaskLock struct {
	// etcd客户端
	kv    clientv3.KV
	lease clientv3.Lease

	lockName   string             // 任务名 + zk + 唯一值
	cancelFunc context.CancelFunc // 用于终止自动续租
	leaseId    clientv3.LeaseID   // 租约ID
	isLocked   bool               // 是否上锁成功
}

// InitTaskLock 初始化一把锁
func InitTaskLock(lockName string, kv clientv3.KV, lease clientv3.Lease) (taskLock *TaskLock) {
	taskLock = &TaskLock{
		kv:       kv,
		lease:    lease,
		lockName: lockName,
	}
	return
}
