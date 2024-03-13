package worker

import (
	"context"
	"fmt"
	"net"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var localIp string

type register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	localIP string
}

var GRegister *register

// 注册节点到etcd： /cron/workers/zk/IP地址 并自动续租
func keepOnline(workPath, zk string) {
	for {
		// 注册路径
		regKey := workPath + zk + "/" + localIp
		fmt.Println(regKey)

		// 创建租约
		leaseGrantResp, err := GRegister.lease.Grant(context.TODO(), 10)
		if err != nil {
			return
		}

		// 自动续租
		keepAliveChan, err := GRegister.lease.KeepAlive(context.TODO(), leaseGrantResp.ID)
		if err != nil {
			return
		}

		cancelCtx, cancelFunc := context.WithCancel(context.TODO())

		// 注册到etcd
		if _, err = GRegister.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			if cancelFunc != nil {
				cancelFunc()
			}
			return
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp := <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					if cancelFunc != nil {
						cancelFunc()
					}
					return
				}
			}
		}
	}
}

// getLocalIP 获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {

	// 获取所有网卡
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr := range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet := addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}
	return
}

func InitRegister(addr []string, workPath, zk string) (err error) {

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

	// 本机IP
	localIp, err = getLocalIP()
	if err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	GRegister = &register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIp,
	}

	// 服务注册
	go keepOnline(workPath, zk)
	return
}
