package model

type OperateCmd interface{}

type Operate struct {
	Timeout   int64      `json:"timeout"`   //毫秒
	ReplySize int        `json:"replySize"` //重试次数
	Cmd       OperateCmd `json:"cmd"`       //命令
}
