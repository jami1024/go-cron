# 介绍

go-cron 是基于golang实现统一定时任务平台。

![流程图](./img/流程图.png)



## 安装
```shell
https://github.com/jami1024/go-cron
```
## 运行
> 需要安装docker-compose后在项目亘目录执行`docker-compose up -d`安装etcd
```shell
cd go-cron
go run cmd/server/main.go 

config.Conf.EtcdConfig.Address
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /version                  --> go-cron/internal/web.InitWeb.func1 (3 handlers)
[GIN-debug] GET    /swagger/*any             --> github.com/swaggo/gin-swagger.CustomWrapHandler.func1 (3 handlers)
[GIN-debug] POST   /tasks/save               --> go-cron/internal/web.(*TaskHandler).Save-fm (3 handlers)

```
在目启动后请访问[后台地址](http://127.0.0.1:8181/swagger/index.html)查询相关接口


