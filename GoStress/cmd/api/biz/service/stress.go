package service

import (
	"context"
	"errors"
	"io"

	"github.com/lizaiganshenmo/GoStress/cmd/api/biz/types/req"
	"github.com/lizaiganshenmo/GoStress/cmd/api/biz/types/resp"
	"github.com/lizaiganshenmo/GoStress/library"
	"github.com/lizaiganshenmo/GoStress/library/utils"
	"github.com/lizaiganshenmo/GoStress/resource"
	"github.com/lizaiganshenmo/GoStress/resource/db"
)

const (
	defaultVerify  = "httpJson"
	defaultTimeout = 30
)

// create task
func CreateTask(ctx context.Context, req *req.CreateTaskReq) (*resp.CreateTaskResp, error) {
	file, err := req.TaskFile.Open()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	curl := utils.Parse(data)

	url := curl.GetURL()
	verify := req.Verify
	if verify == "" {
		verify = defaultVerify
	}
	taskID := resource.TaskSF.NextVal()
	err = db.CreateTask(ctx, &db.StressTask{
		TaskID:      taskID,
		TargetQPS:   req.TargetQPS,
		URL:         url,
		Protocol:    library.GetURLProtocol(url),
		Method:      curl.GetMethod(),
		Headers:     curl.GetHeadersStr(),
		Body:        curl.GetBody(),
		Verify:      verify,
		Timeout:     defaultTimeout,
		UseHTTP2:    req.UseHTTP2,
		KeepAlive:   req.KeepAlive,
		Description: req.Description,
	})

	return &resp.CreateTaskResp{TaskID: taskID}, err

}

// StartTask
func StartTask(ctx context.Context, req *req.StartTaskReq) error {
	// 查询该任务是否是未执行
	res, err := db.GetTaskByTaskID(ctx, req.TaskID, db.NoExec)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return errors.New("not exist noexec task")
	}

	err = db.UpdateTask(ctx, &db.StressTask{
		TaskID: req.TaskID,
		Status: db.WaitExec,
	})
	return err

}

// Stop Task
func StopTask(ctx context.Context, req *req.StopTaskReq) error {
	// 查询该任务是否是执行中
	res, err := db.GetTaskByTaskID(ctx, req.TaskID, db.Execing)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return errors.New("not exist execing task")
	}

	err = db.UpdateTask(ctx, &db.StressTask{
		TaskID: req.TaskID,
		Status: db.WaitStop,
	})
	return err

}

// Stop Task
func ChangeTaskQPS(ctx context.Context, req *req.ChangeTaskQPSReq) error {
	// 查询该任务是否是执行中
	res, err := db.GetTaskByTaskID(ctx, req.TaskID, db.Execing)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return errors.New("not exist execing task")
	}

	err = db.UpdateTask(ctx, &db.StressTask{
		TaskID:    req.TaskID,
		TargetQPS: req.TargetQPS,
		Status:    db.WaitChangeQPS,
	})
	return err

}
