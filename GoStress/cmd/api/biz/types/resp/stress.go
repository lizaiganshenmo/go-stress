package resp

type CreateTaskResp struct {
	TaskID int64 `form:"task_id" json:"task_id"`
}
