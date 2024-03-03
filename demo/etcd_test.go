package demo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.etcd.io/etcd/client/v3"
)

func TestEtcd(t *testing.T) {

	// 客户端配置
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:12379"},
		DialTimeout: 5 * time.Second,
	}

	// 建立连接
	cli, err := clientv3.New(config)
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(cli)
	if err != nil {
		fmt.Println(err)
		return
	}
	//建立用于读写etcd键值对
	kv := clientv3.NewKV(cli)

	//写入
	putRes, err := kv.Put(
		context.TODO(),
		"/cron/zk/job33",
		"hello 33",
		clientv3.WithPrevKV())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Revision:", putRes.Header.Revision)
	if putRes.PrevKv != nil {
		// 存在覆盖value
		fmt.Println("已把内容:", string(putRes.PrevKv.Value), "覆盖")
	}

	//读取, 参数clientv3.WithPrefix() 读取/cron/zk/为前缀的所有key

	getRes, err := kv.Get(
		context.TODO(),
		"/cron/zk",
		clientv3.WithPrefix(),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(getRes.Kvs, getRes.Count)

	// 删除
	delRes, err := kv.Delete(
		context.TODO(),
		"/cron/zk/job33",
		clientv3.WithPrevKV(),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 被删除之前的value是什么
	if len(delRes.PrevKvs) != 0 {
		for _, keypair := range delRes.PrevKvs {
			fmt.Println("删除了:", string(keypair.Key), string(keypair.Value))
		}
	}
}

func TestEtcdLease(t *testing.T) {
	// 测试etcd租期
	// 客户端配置
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:12379"},
		DialTimeout: 5 * time.Second,
	}

	// 建立连接
	cli, err := clientv3.New(config)
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(cli)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 创建一个一个lease（租约）对象
	lease := clientv3.NewLease(cli)

	// 申请一个10秒的租约
	leaseGrantResp, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 拿到租约的ID
	leaseId := leaseGrantResp.ID
	// 5秒后会取消自动续租
	//ctx, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
	//自动续租
	keepRespChan, err := lease.KeepAlive(context.TODO(), leaseId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 处理续约应答的协程
	go func() {
		for {
			select {
			case keepResp := <-keepRespChan:
				if keepRespChan == nil {
					fmt.Println("租约已经失效了")
					//goto END
					break
				} else { // 每秒会续租一次, 所以就会受到一次应答
					fmt.Println("收到自动续租应答:", keepResp.ID)
				}
			}
		}
		//END:
	}()

	// 获得kv对象
	kv := clientv3.NewKV(cli)
	// Put一个KV, 让它与租约关联起来, 从而实现10秒后自动过期
	putResp, err := kv.Put(context.TODO(), "/cron/lock/job1", "", clientv3.WithLease(leaseId))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("写入成功:", putResp.Header.Revision)

	// 定时的看一下key过期了没有
	for {
		getResp, err := kv.Get(context.TODO(), "/cron/lock/job1")
		if err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("kv过期了")
			break
		}
		fmt.Println("还没过期:", getResp.Kvs)
		time.Sleep(2 * time.Second)
	}
}
