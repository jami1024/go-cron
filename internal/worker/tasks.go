package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-cron/internal/domain"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var GWorkerTaskMgr *domain.WorkerTaskMgr
var keyName string
var keyUniqueCode string
var keyZk string

// unpackJob 反序列化定时任务
func unpackJob(value []byte) (ret domain.Task, err error) {

	if err = json.Unmarshal(value, &ret); err != nil {
		return
	}
	return
}

// TaskChangeEvent 任务变化事件有2种：1）更新任务 2）删除任务
func TaskChangeEvent(eventType int, task *domain.Task) (taskEvent *domain.TaskEvent) {
	return &domain.TaskEvent{
		EventType: eventType,
		Task:      task,
	}
}

// 监听任务变化
func watchJobs(keyPath, zk string) (err error) {

	// 1, get一下/cron/jobs/zk/目录下的所有任务，并且获知当前集群的revision
	dir := keyPath + zk
	fmt.Println("任务dir", dir)
	getResp, err := GWorkerTaskMgr.Kv.Get(context.TODO(), dir, clientv3.WithPrefix())

	if err != nil {
		return
	}

	// 当前有哪些任务
	for _, keypair := range getResp.Kvs {
		// 反序列化json得到Task
		task, err := unpackJob(keypair.Value)
		if err == nil {

			fmt.Println("已存在任务", task)

			// 构建一个已存在Event

			taskEvent := TaskChangeEvent(1, &task)
			fmt.Println("已存在Event", taskEvent)

			// 发送任务给调度器
			G_scheduler.PushTaskEvent(taskEvent)
			fmt.Println("发送已存在任务给调度器完毕")
		}

	}

	// 2, 从该revision向后监听变化事件
	go func() { // 监听协程
		// 从GET时刻的后续版本开始监听变化
		watchStartRevision := getResp.Header.Revision + 1
		// 监听/cron/jobs/zk/目录的后续变化
		watchChan := GWorkerTaskMgr.Watcher.Watch(context.TODO(), dir, clientv3.WithRev(watchStartRevision),
			clientv3.WithPrefix())
		// 处理监听事件
		for watchResp := range watchChan {
			for _, watchEvent := range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 任务保存事件
					task, err := unpackJob(watchEvent.Kv.Value)
					if err != nil {
						// 解析错误跳过，定时任务的格式异常
						continue
					}
					fmt.Println("新增/编辑任务", task)
					// 构建一个更新Event
					taskEvent := TaskChangeEvent(1, &task)
					// 发送任务给调度器
					G_scheduler.PushTaskEvent(taskEvent)
					fmt.Println("更新Event", taskEvent)

				case mvccpb.DELETE: // 任务被删除了
					// Delete /cron/jobs/zk/xx
					fmt.Println("删除任务", string(watchEvent.Kv.Key))
					taskKey := string(watchEvent.Kv.Key)
					parts := strings.Split(taskKey, "/")

					if len(parts) != 5 {
						lastPart := parts[len(parts)-1]
						//fmt.Println(lastPart)
						keyZk = parts[3]
						keyName = strings.Split(lastPart, "_")[0]
						keyUniqueCode = strings.Split(lastPart, "_")[1]
					} else {
						continue
					}
					// 删除任务需要任务名称和唯一值和对应中控
					task := &domain.Task{Name: keyName, UniqueCode: keyUniqueCode, Zk: keyZk}
					// 构建一个删除Event
					taskEvent := TaskChangeEvent(2, task)
					// 发送任务给调度器
					G_scheduler.PushTaskEvent(taskEvent)
					fmt.Println("删除Event", taskEvent)

				}
			}
		}
	}()
	return
}

// InitTaskMgr 初始化worker的任务管理
func InitTaskMgr(addr []string, keyPath, zk string) (err error) {

	// 初始化配置
	config := clientv3.Config{
		Endpoints:   addr,                   // 集群地址
		DialTimeout: 300 * time.Millisecond, // 连接超时
	}

	// 建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)

	// 赋值单例
	GWorkerTaskMgr = &domain.WorkerTaskMgr{
		Client:  client,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	// 启动任务监听
	err = watchJobs(keyPath, zk)
	if err != nil {
		return err
	}
	return
}
