package worker

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	"go-cron/internal/domain"
)

// Executor 任务执行器
type Executor struct {
}

var (
	G_executor *Executor
)

// ExecuteTask 执行一个任务
func (executor *Executor) ExecuteTask(info *domain.TaskExecuteInfo) {

	// 任务结果
	result := &domain.TaskExecuteResult{
		ExecuteInfo: info,
		Output:      make([]byte, 0),
	}
	// 记录任务开始时间
	result.StartTime = time.Now()

	// 随机睡眠(0~1s)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	// todo 只能执行bash脚本和命令、后续要支持python脚本
	// 执行shell命令
	cmd := exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Task.Command)
	// 执行并捕获输出
	output, err := cmd.CombinedOutput()
	// 记录任务结束时间
	result.EndTime = time.Now()
	result.Output = output
	result.Err = err
	fmt.Println(output, err)
	// 任务执行完成后，把执行的结果返回给Scheduler，Scheduler会从executingTable中删除掉执行记录
	G_scheduler.PushTaskResult(result)

}

// InitExecutor 初始化执行器
func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
