package worker

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
	"go-cron/internal/domain"
)

// Scheduler 任务调度
type Scheduler struct {
	taskEventChan      chan *domain.TaskEvent              //  etcd任务事件队列,新增｜删除
	taskPlanTable      map[string]*domain.TaskSchedulePlan // 任务调度计划表
	taskExecutingTable map[string]*domain.TaskExecuteInfo  // 任务执行表
	taskResultChan     chan *domain.TaskExecuteResult      // 任务结果队列
}

var (
	G_scheduler *Scheduler
)

// handleTaskEvent 处理任务事件
func (scheduler *Scheduler) handleTaskEvent(taskEvent *domain.TaskEvent) {

	switch taskEvent.EventType {
	case 1: // 保存任务事件
		taskSchedulePlan, err := BuildJobSchedulePlan(taskEvent.Task)
		if err != nil {
			return
		}
		// name + 唯一值
		key := taskEvent.Task.Name + "_" + taskEvent.Task.UniqueCode
		fmt.Println("新增前定时任务列表", key, scheduler.taskPlanTable)
		scheduler.taskPlanTable[key] = taskSchedulePlan
		fmt.Println("新增后定时任务列表", key, scheduler.taskPlanTable)
	case 2: // 删除任务事件
		// name + 唯一值
		key := taskEvent.Task.Name + "_" + taskEvent.Task.UniqueCode
		fmt.Println("删除前定时任务列表", key, scheduler.taskPlanTable)
		// 检查是否存在
		_, taskExisted := scheduler.taskPlanTable[key]
		if taskExisted {
			delete(scheduler.taskPlanTable, key)
		}
		fmt.Println("删除后定时任务列表", key, scheduler.taskPlanTable)
	}
}

// handleTaskResult 处理任务结果
func (scheduler *Scheduler) handleTaskResult(result *domain.TaskExecuteResult) {
	// 删除执行状态
	key := result.ExecuteInfo.Task.Name + "_" + result.ExecuteInfo.Task.UniqueCode
	//fmt.Println("任务", key)
	delete(scheduler.taskExecutingTable, key)
	fmt.Println("执行完成的任务从内存中删除:", result.ExecuteInfo.Task.Name, string(result.Output), result.Err)

}

// TrySchedule 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var nearTime *time.Time
	// 如果任务表为空话，下次调度间隔为1秒
	if len(scheduler.taskPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	// 当前时间
	now := time.Now()

	// 遍历所有任务
	fmt.Println("准备遍历计划任务", scheduler.taskPlanTable)
	for _, taskPlan := range scheduler.taskPlanTable {

		// 小于或者等于当前时间，证明任务到期，要执行
		if taskPlan.NextTime.Before(now) || taskPlan.NextTime.Equal(now) {

			// 启动任务，注意如果上一次还没有结束本次会再次执行
			scheduler.tryStartTask(taskPlan)

			// 更新下次执行时间
			taskPlan.NextTime = taskPlan.Expr.Next(now)
		}

		// 统计最近一个要过期的任务时间，用于下次调度间隔
		if nearTime == nil || taskPlan.NextTime.Before(*nearTime) {
			nearTime = &taskPlan.NextTime
		}
	}
	// 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
	scheduleAfter = (*nearTime).Sub(now)
	return
}

// tryStartTask 执行任务
func (scheduler *Scheduler) tryStartTask(taskPlan *domain.TaskSchedulePlan) {
	// 取消该步骤如果任务正在执行，跳过本次调度
	key := taskPlan.Task.Name + "_" + taskPlan.Task.UniqueCode
	//_, taskExecuting := scheduler.taskExecutingTable[key]
	//if taskExecuting {
	//	fmt.Println("尚未退出,跳过执行:", key)
	//	return
	//}
	// 构建执行状态信息
	taskExecuteInfo := domain.BuildTaskExecuteInfo(taskPlan)

	// 保存执行状态
	scheduler.taskExecutingTable[key] = taskExecuteInfo

	// 执行任务
	fmt.Println("触发执行任务:", taskExecuteInfo.Task.Name, taskExecuteInfo.PlanTime,
		taskExecuteInfo.RealTime)
	G_executor.ExecuteTask(taskExecuteInfo)
}

// 调度协程
func (scheduler *Scheduler) scheduleLoop() {
	// 等待
	// 获取下一次调度间隔，初始化一次(1秒)
	scheduleAfter := scheduler.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer := time.NewTimer(scheduleAfter)
	for {
		select {
		case taskEvent := <-scheduler.taskEventChan: //监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			fmt.Println("监听任务变化事件", taskEvent.Task.Name, taskEvent.EventType)
			scheduler.handleTaskEvent(taskEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
		case taskResult := <-scheduler.taskResultChan: // 监听任务执行结果
			scheduler.handleTaskResult(taskResult)
		}
		// 调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// PushTaskEvent 推送任务变化事件
func (scheduler *Scheduler) PushTaskEvent(taskEvent *domain.TaskEvent) {
	scheduler.taskEventChan <- taskEvent
}

// BuildJobSchedulePlan 构造任务执行计划
func BuildJobSchedulePlan(task *domain.Task) (taskSchedulePlan *domain.TaskSchedulePlan, err error) {

	// 解析JOB的cron表达式
	expr, err := cronexpr.Parse(task.CronExpr)
	if err != nil {
		return
	}

	// 生成任务调度计划对象
	taskSchedulePlan = &domain.TaskSchedulePlan{
		Task:     task,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

// PushTaskResult 回传任务执行结果
func (scheduler *Scheduler) PushTaskResult(taskResult *domain.TaskExecuteResult) {
	scheduler.taskResultChan <- taskResult
}

// InitScheduler 初始化调度器
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		taskEventChan:      make(chan *domain.TaskEvent, 10000),
		taskPlanTable:      make(map[string]*domain.TaskSchedulePlan),
		taskExecutingTable: make(map[string]*domain.TaskExecuteInfo),
		taskResultChan:     make(chan *domain.TaskExecuteResult, 10000),
	}
	// 启动调度协程
	go G_scheduler.scheduleLoop()
	fmt.Printf("scheduleLoop完毕")
	return
}
