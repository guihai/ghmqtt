package server

import (
	"github.com/guihai/ghmqtt/mqtt311/proto"
)

type ImplBaseRouter interface {
	//
	PreHandle(request *Request)
	Handle(request *Request)
	PostHandle(request *Request)
}

/*
设定基础路由，业务路由继承此路由
*/
type BaseRouter struct {
}

// 前置方法
func (s *BaseRouter) PreHandle(request *Request) {

}

// 业务处理方法
func (s *BaseRouter) Handle(request *Request) {

}

// 后置方法
func (s *BaseRouter) PostHandle(request *Request) {

}

//  定义一个方法别名，本方法在连接时候使用，仅使用一次
type ConnectVerifyFUNC func(*proto.CONNECTProtocol) uint8
