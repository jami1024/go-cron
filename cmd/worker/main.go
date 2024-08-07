package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"go-cron/config"
	"go-cron/internal/worker"
	"go-cron/pkg/logger"

	"go.uber.org/zap"
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
	// 1. 加载配置
	if err := config.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	// 2. 初始化日志
	err := logger.WorkerInit(config.Conf.WorkerLogConfig)
	if err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
	}
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			fmt.Printf("sync log failed, err:%v\n", err)
			return
		}
	}(zap.L())
	zap.L().Debug("logger init success...")

	// 初始化命令行参数
	initArgs()

	// 服务注册
	err = worker.InitRegister(GConfig.EtcdAddress, GConfig.WorkerPath, GConfig.Zk)
	if err != nil {
		zap.L().Sugar().Errorf("worker注册失败,%s", err.Error())
		return
	}
	zap.L().Sugar().Info("worker注册成功")

	// 初始化执行器
	err = worker.InitExecutor()
	if err != nil {
		zap.L().Sugar().Errorf("worker初始化执行器,%s", err.Error())
		return
	}
	zap.L().Sugar().Info("worker初始化执行器成功")

	// 初始化调度器
	err = worker.InitScheduler()
	if err != nil {
		zap.L().Sugar().Errorf("worker初始化调度器,%s", err.Error())
		return
	}
	zap.L().Sugar().Info("worker初始化调度器成功")

	// 初始化任务管理器
	err = worker.InitTaskMgr(GConfig.EtcdAddress, GConfig.KeyPath, GConfig.Zk)
	if err != nil {
		zap.L().Sugar().Errorf("worker初始化任务管理器,%s", err.Error())
		return
	}
	zap.L().Sugar().Info("worker初始化任务管理器成功")

	// 正常退出
	for {
		time.Sleep(1 * time.Second)
	}
}
