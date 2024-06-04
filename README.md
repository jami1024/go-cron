# 介绍

go-cron 是基于golang实现统一定时任务平台。

![流程图](./img/流程图.png)



## 安装
```shell
https://github.com/jami1024/go-cron
```

## 修改配置
```
# 编辑config/config.json
# 根据实际情况修改etcd.address
```
## 运行server
> 需要安装docker-compose后在项目根目录执行`docker-compose up -d`安装etcd
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

## docker compose
```
# 构建go-cron docker镜像
docker build -t go-cron-server:latest .

# 运行docker compose
docker compose up -d

注意：docker compose集成etcd、go-cron-web(前端程序)、go-cron(后端程序)
```

## 运行worker
编辑config/worker.json，根据实际情况修改etcd.address、etcd.zk。

**etcd.zk对照表：**
|前端显示|worker显示|
|---|---|
|国内|bjzk|
|日本|jpzk|
|欧美|uszk|
|韩国|krzk|

```
# 运行worker
go run cmd/worker/main.go
```
## 前端项目
[go-cron-web](https://github.com/jami1024/go-cron-web)


