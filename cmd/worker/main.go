package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"go-cron/internal/worker"
)

var GConfig *Config

// Config worker程序配置
type Config struct {
	EtcdAddress []string `json:"etcdAddress"`
	Zk          string   `json:"zk"`
	KeyPath     string   `json:"keyPath"`
	WorkerPath  string   `json:"workerPath"`
}

// 解析命令行参数
func initArgs() {
	// worker -config ./worker.json
	// worker -h
	jsonFilePath := flag.String("config", "", "the path to the JSON file")
	flag.Parse()
	// 检查是否提供了文件路径
	if *jsonFilePath == "" {
		fmt.Println("Please specify the path to the json file using -config option")
		return
	}
	// 1, 把配置文件读进来
	fileContent, err := os.ReadFile(*jsonFilePath)
	if err != nil {
		fmt.Printf("Error reading JSON file: %s\n", err)
		return
	}

	// 2, 做JSON反序列化
	var config Config
	if err = json.Unmarshal(fileContent, &config); err != nil {
		return
	}
	GConfig = &config
	return
}

func main() {
	// 初始化命令行参数
	initArgs()
	// 服务注册
	err := worker.InitRegister(GConfig.EtcdAddress, GConfig.WorkerPath, GConfig.Zk)
	if err != nil {
		fmt.Printf("worker注册失败,%s", err.Error())
		return
	}
	fmt.Println("worker注册成功")

	// 初始化任务管理器
	err = worker.InitTaskMgr(GConfig.EtcdAddress, GConfig.KeyPath, GConfig.Zk)
	if err != nil {
		fmt.Printf("worker初始化任务管理器,%s", err.Error())
		return
	}
	fmt.Println("worker初始化任务管理器成功")

	// 正常退出
	for {
		time.Sleep(1 * time.Second)
	}
}
