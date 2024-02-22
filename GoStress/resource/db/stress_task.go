package db

import (
	"context"

	"github.com/lizaiganshenmo/GoStress/resource"
	"gorm.io/gorm"
)

const (
	// 任务status
	NoExec = iota
	WaitExec
	Execing
	WaitChangeQPS
	WaitStop
	HasExeced

	StressTaskTableName = "stress_task"
)

type StressTask struct {
	gorm.Model
	TaskID      int64  `json:"task_id" gorm:"column:task_id"`
	TargetQPS   int    `json:"target_qps" gorm:"column:target_qps"`
	Status      int    `json:"status" gorm:"column:status"`
	URL         string `json:"url" gorm:"column:url"`
	Protocol    string `json:"protocol" gorm:"column:protocol"`
	Method      string `json:"method" gorm:"column:method"`
	Headers     string `json:"headers" gorm:"column:headers"` // db中以json字符串形式存储
	Body        string `json:"body" gorm:"column:body"`
	Verify      string `json:"verify" gorm:"column:verify"`
	Timeout     int    `json:"timeout" gorm:"column:timeout"`
	UseHTTP2    bool   `json:"use_http2" gorm:"column:use_http2"`
	KeepAlive   bool   `json:"keepalive" gorm:"column:keepalive"`
	Description string `json:"Description" gorm:"column:description"`
}

func (s *StressTask) TableName() string {
	return StressTaskTableName
}

// 查询任务
func GetTaskByStatuss(ctx context.Context, status []int) (tasks []*StressTask, err error) {
	db := resource.MySQLStressDB.WithContext(ctx).Model(&StressTask{})
	err = db.Where("status in ?", status).Find(&tasks).Error

	return
}

// 查询任务-by taskid
func GetTaskByTaskID(ctx context.Context, taskID int64, status int) (tasks []*StressTask, err error) {
	db := resource.MySQLStressDB.WithContext(ctx).Model(&StressTask{})
	err = db.Where("task_id = ? and status = ?", taskID, status).Find(&tasks).Error

	return
}

// create stress task
func CreateTask(ctx context.Context, taskInfo *StressTask) error {
	return resource.MySQLStressDB.WithContext(ctx).Model(&StressTask{}).Create(taskInfo).Error
}

// update
func UpdateTask(ctx context.Context, taskInfo *StressTask) error {
	return resource.MySQLStressDB.WithContext(ctx).Model(&StressTask{}).Where("task_id = ?", taskInfo.TaskID).Updates(taskInfo).Error
}

// delete
func DeleteTask(ctx context.Context, taskInfo *StressTask) error {
	return resource.MySQLStressDB.WithContext(ctx).Model(&StressTask{}).Where("task_id = ?", taskInfo.TaskID).Delete(taskInfo).Error
}
