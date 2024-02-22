package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/lizaiganshenmo/GoStress/cmd/api/biz/service"
	"github.com/lizaiganshenmo/GoStress/cmd/api/biz/types/req"
	"github.com/lizaiganshenmo/GoStress/library/errno"
)

// create task
func CreateTask(ctx context.Context, c *app.RequestContext) {

	var req req.CreateTaskReq
	if err := c.Bind(&req); err != nil {
		SendResponse(c, errno.ParamErr, nil)
		return
	}

	resp, err := service.CreateTask(ctx, &req)
	if err != nil {
		SendResponse(c, err, nil)
		hlog.CtxWarnf(ctx, "service.CreateTask fail. err: %+v. req:%+v", err, req)
		return
	}

	SendResponse(c, errno.Success, resp)

}

// start task
func StartTask(ctx context.Context, c *app.RequestContext) {
	var req req.StartTaskReq
	if err := c.Bind(&req); err != nil {
		SendResponse(c, errno.ParamErr, nil)
		return
	}

	err := service.StartTask(ctx, &req)
	if err != nil {
		SendResponse(c, err, nil)
		hlog.CtxWarnf(ctx, "service.StartTask fail. err: %+v. req:%+v", err, req)
		return
	}

	SendResponse(c, errno.Success, nil)

}

// stop task
func StopTask(ctx context.Context, c *app.RequestContext) {
	var req req.StopTaskReq
	if err := c.Bind(&req); err != nil {
		SendResponse(c, errno.ParamErr, nil)
		return
	}

	err := service.StopTask(ctx, &req)
	if err != nil {
		SendResponse(c, err, nil)
		hlog.CtxWarnf(ctx, "service.StopTask fail. err: %+v. req:%+v", err, req)
		return
	}

	SendResponse(c, errno.Success, nil)

}

// change task qps
func ChangeTaskQPS(ctx context.Context, c *app.RequestContext) {
	var req req.ChangeTaskQPSReq
	if err := c.Bind(&req); err != nil {
		SendResponse(c, errno.ParamErr, nil)
		return
	}

	err := service.ChangeTaskQPS(ctx, &req)
	if err != nil {
		SendResponse(c, err, nil)
		hlog.CtxWarnf(ctx, "service.ChangeTaskQPS fail. err: %+v. req:%+v", err, req)
		return
	}

	SendResponse(c, errno.Success, nil)

}
