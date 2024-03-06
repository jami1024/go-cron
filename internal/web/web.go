package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-cron/config"
	_ "go-cron/docs"
	"go-cron/internal/domain"
	"go-cron/internal/service"
	"go-cron/pkg/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func InitWeb(zapL *zap.Logger) *gin.Engine {
	// 初始化etcd
	GTaskmgr, err := initEtcd(config.Conf.EtcdConfig.Address)
	if err != nil {
		zapL.Error(fmt.Sprintf("初始化etcd失败", err))
	}

	// gin 模式
	if config.Conf.Mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	server := gin.New()
	server.Use(logger.GinLogger(), logger.GinRecovery(true))

	server.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Conf.Version)
	})
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 任务路由
	//userHandler := web.NewUserHandler(userService, codeService, handler)
	// service.NewUserService(userRepository, loggerV1)
	taskHandler := NewTaskHandler(service.NewTaskService(zapL, GTaskmgr), zapL)
	taskHandler.RegisterRoutes(server)
	return server
}

func initEtcd(addr []string) (GTaskmgr *domain.TaskMgr, err error) {
	// 初始化配置
	etcdConfig := clientv3.Config{
		Endpoints:   addr,                                   // 集群地址
		DialTimeout: time.Duration(5000) * time.Millisecond, // 连接超时
	}
	// 建立连接
	client, err := clientv3.New(etcdConfig)
	//defer func(cli *clientv3.Client) {
	//	err := cli.Close()
	//	if err != nil {
	//		zap.L().Info(fmt.Sprintf("初始化etcd失败,%v", err.Error()))
	//		return
	//	}
	//}(client)
	if err != nil {
		return nil, err
	}
	// 得到KV和Lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	// 赋值单例
	GTaskmgr = &domain.TaskMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	//zap.L().Info(fmt.Sprintf("链接ETCD完毕,%v", GTaskmgr))
	return GTaskmgr, err
}
