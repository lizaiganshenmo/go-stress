package req

import "mime/multipart"

type CreateTaskReq struct {
	TaskFile    *multipart.FileHeader `form:"task_file" json:"task_file"`
	TargetQPS   int                   `form:"target_qps" json:"target_qps"`
	Verify      string                `form:"verify" json:"verify"`
	UseHTTP2    bool                  `form:"use_http2" json:"use_http2"`
	KeepAlive   bool                  `form:"keepalive" json:"keepalive"`
	Description string                `form:"description" json:"description"`
}

type StartTaskReq struct {
	TaskID int64 `form:"task_id" json:"task_id"`
}

type StopTaskReq struct {
	TaskID int64 `form:"task_id" json:"task_id"`
}

type ChangeTaskQPSReq struct {
	TaskID    int64 `form:"task_id" json:"task_id"`
	TargetQPS int   `form:"target_qps" json:"target_qps"`
}
