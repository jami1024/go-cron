package worker

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
	"go-cron/internal/domain"
)

// Scheduler 任务调度
type Scheduler struct {
	taskEventChan chan *domain.TaskEvent              //  etcd任务事件队列,新增｜删除
	taskPlanTable map[string]*domain.TaskSchedulePlan // 任务调度计划表

}

var (
	G_scheduler *Scheduler
)

// 处理任务事件
func (scheduler *Scheduler) handleTaskEvent(taskEvent *domain.TaskEvent) {

	switch taskEvent.EventType {
	case 1: // 保存任务事件
		taskSchedulePlan, err := BuildJobSchedulePlan(taskEvent.Task)
		if err != nil {
			return
		}
		// name + 唯一值
		key := taskEvent.Task.Name + "_" + taskEvent.Task.UniqueCode
		fmt.Println("1 key", key)
		scheduler.taskPlanTable[key] = taskSchedulePlan
	case 2: // 删除任务事件
		// name + 唯一值
		key := taskEvent.Task.Name + "_" + taskEvent.Task.UniqueCode
		fmt.Println("2 key", key)
		// 检查是否存在
		_, taskExisted := scheduler.taskPlanTable[key]

		if taskExisted {
			delete(scheduler.taskPlanTable, key)
		}
	}
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
	for _, taskPlan := range scheduler.taskPlanTable {

		// 小于或者等于当前时间，证明任务到期，要执行
		if taskPlan.NextTime.Before(now) || taskPlan.NextTime.Equal(now) {

			// todo 尝试启动任务，注意如果上一次还没有结束本次要不要执行？
			fmt.Println("执行任务", taskPlan.Task.Name, taskPlan.Task.Zk, taskPlan.Task.Command,
				taskPlan.Task.CronExpr)
			//scheduler.TryStartJob(jobPlan)

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
			fmt.Println(taskEvent)
			scheduler.handleTaskEvent(taskEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
			//case jobResult = <-scheduler.jobResultChan: // 监听任务执行结果
			//scheduler.handleJobResult(jobResult)
		}
		// 调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		// 重置调度间隔

		fmt.Println("间隔", scheduleAfter)
		scheduleTimer.Reset(scheduleAfter)
	}
}

// PushTaskEvent 推送任务变化事件
func (scheduler *Scheduler) PushTaskEvent(taskEvent *domain.TaskEvent) {
	fmt.Println("1111")
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

// InitScheduler 初始化调度器
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		taskEventChan: make(chan *domain.TaskEvent, 1000),
		taskPlanTable: make(map[string]*domain.TaskSchedulePlan),
	}
	// 启动调度协程
	go G_scheduler.scheduleLoop()
	fmt.Printf("scheduleLoop完毕")
	return
}
