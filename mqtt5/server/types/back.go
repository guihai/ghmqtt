package types

import (
	"github.com/guihai/ghmqtt/utils"
)

// 响应结构体
type Response struct {
	Code int64       `json:"Code"`
	Msg  string      `json:"Msg"`
	Data interface{} `json:"Data,omitempty"`
}

func NewResponse() *Response {
	return &Response{
		Code: utils.RECODE_UNKNOWERR, // 初始为未知错误
		Msg:  utils.MsgText(utils.RECODE_UNKNOWERR),
		Data: nil,
	}
}

type PageData struct {
	Page     int64       `json:"Page,omitempty"`
	PageSize int64       `json:"PageSize,omitempty"`
	Search   interface{} `json:"Search"`
	Count    int64       `json:"Count"`
	Data     interface{} `json:"Data,omitempty"`
	DataList interface{} `json:"DataList"`
}

func NewPageData() *PageData {
	return &PageData{
		Page:     0,
		PageSize: 0,
		Search:   nil,
		Count:    0,
		DataList: []int64{},
		Data:     nil,
	}
}

type PageReq struct {
	Page     int64 `json:"Page"`
	PageSize int64 `json:"PageSize"`
}
