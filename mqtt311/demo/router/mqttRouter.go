package router

import (
	"fmt"
	"github.com/guihai/ghmqtt/mqtt311/proto"
	"github.com/guihai/ghmqtt/mqtt311/server"
)

// 实现链接验证方法
func CheckConn(protocol *proto.CONNECTProtocol) uint8 {
	//fmt.Println("这是CONNECTProtocolRouter  CheckConn")

	//if protocol.UserName == "" {
	//	return proto.Refused_i_r
	//}
	return 0
}

type CONNECTRouter struct {
	*server.BaseRouter
}

// 业务处理方法
func (s *CONNECTRouter) Handle(request *server.Request) {
	//fmt.Println("这是CONNECTProtocolRouter  Handle")
	//
	//fmt.Println("proto.HeaderFlag", request.Proto.(*proto.CONNECTProtocol).HeaderFlag)
}

//默认的断开链接路由器
type DISCONNECTRouter struct {
	*server.BaseRouter
}

// 默认断开路由方法
func (s *DISCONNECTRouter) Handle(request *server.Request) {
	// 断开链接
	request.ConnStop()
}

//默认心跳协议路由
type PINGREQRouter struct {
	*server.BaseRouter
}

// 默认断开路由方法
func (s *PINGREQRouter) Handle(request *server.Request) {
	// 发送 心跳响应
	// 制作响应结构体，发给响应

	p := &proto.PINGRESPProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PINGRESP,
			MsgLen:     0,
		},
	}

	request.SendRES(p)
}

//默认订阅协议路由
type SUBSCRIBERouter struct {
	*server.BaseRouter
}

func (s *SUBSCRIBERouter) Handle(request *server.Request) {
	//  处理订阅结果，返回响应

	// todo 架构实现协议，业务只返回订阅成功标识
	sp := request.GetProto().(*proto.SUBSCRIBEProtocol)

	// 订阅主题
	for _, filter := range sp.TopicFilterList {
		// 订阅主题
		request.SubTopic(filter.FilterName)
	}

	p := &proto.SUBACKProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.SUBACK,
			MsgLen:     uint32(2 + len(sp.TopicFilterList)), // 标识符 2 字节，返回code len(sp.TopicFilterList)
			Data:       nil,
		},
		// 和订阅协议相同
		PacketIdentifier: sp.PacketIdentifier,
		// 全部返回订阅成功
		ReturnCodeList: make([]byte, len(sp.TopicFilterList)),
	}

	request.SendRES(p)
}

//默认 取消订阅协议路由
type UNSUBSCRIBERouter struct {
	*server.BaseRouter
}

func (s *UNSUBSCRIBERouter) Handle(request *server.Request) {

	sp := request.GetProto().(*proto.UNSUBSCRIBEProtocol)

	// 取消订阅主题
	for _, filter := range sp.TopicFilterList {
		// 取消订阅主题
		request.UnSubTopic(filter.FilterName)
	}

	p := &proto.UNSUBACKProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.UNSUBACK,
			MsgLen:     uint32(2), // 标识符 2 字节
			Data:       nil,
		},
		// 和订阅协议相同
		PacketIdentifier: sp.PacketIdentifier,
	}

	request.SendRES(p)
}

//默认 发布协议
type PUBLISHRouter struct {
	*server.BaseRouter
}

func (s *PUBLISHRouter) Handle(request *server.Request) {

	sp := request.GetProto().(*proto.PUBLISHProtocol)

	/*
			Qos0
			Qos1
			Qos2
			三种情况处理

		retain 保留信息 两种情况处理
	*/

	// 发布的数据 进入 主题管理器
	// 版本1 直接开 主题接收发送协程
	//go request.MsgIn(sp.TopicName, sp.Payload)
	// 版本2 进入协程池
	go request.MsgInPool(sp)

	// 保留信息处理 暂定协程
	if sp.Retain {
		go request.SetRetainMsg(sp.TopicName, sp.Payload)
	}

	// 	Qos 处理
	switch sp.Qos {
	case proto.QoS1:

		// 返回 puback
		p := &proto.PUBACKProtocol{
			Fixed: &proto.Fixed{
				HeaderFlag: proto.PUBACK,
				MsgLen:     2,
				Data:       nil,
			},
			// 标识符 和 publish 一样
			PacketIdentifier: sp.PacketIdentifier,
		}

		//
		request.SendRES(p)

	case proto.QoS2:
		/*
				1，验证标识符是否存在，存在的有其他信息在使用，不能存储
			2，存储标识符
			3，返回 PUBRECProtocol 协议
		*/

		if request.GetQos2ID(sp.MsgId) {
			// 存在返回真
			// 重复的标识符，不能加入map 要关闭链接
			fmt.Println("重复的标识符，不能加入map 要关闭链接", request.GetConnClientID())

			request.ConnStop()
		}

		request.SetQos2ID(sp.MsgId)

		p := &proto.PUBRECProtocol{
			Fixed: &proto.Fixed{
				HeaderFlag: proto.PUBREC,
				MsgLen:     2,
				Data:       nil,
			},
			PacketIdentifier: sp.PacketIdentifier,
			MsgId:            sp.MsgId,
		}
		request.SendRES(p)

	default:
		// 默认无响应

	}

}

//默认 发布消息响应
type PUBACKRouter struct {
	*server.BaseRouter
}

/*
业务方可以解析 标识符合发布的标识符对应
*/
func (s *PUBACKRouter) Handle(request *server.Request) {
	sp := request.GetProto().(*proto.PUBACKProtocol)

	fmt.Println("puback msg_id === ", sp.MsgId)
}

//默认 释放消息协议
type PUBRELRouter struct {
	*server.BaseRouter
}

/*
1， 移出 QosID
2，返回响应
*/
func (s *PUBRELRouter) Handle(request *server.Request) {
	sp := request.GetProto().(*proto.PUBRELProtocol)

	// 移出 标识id
	request.RemoveQos2ID(sp.MsgId)

	// 需要返回PUBCOMP

	p := &proto.PUBCOMPProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PUBCOMP,
			MsgLen:     2,
			Data:       nil,
		},
		PacketIdentifier: sp.PacketIdentifier,
		MsgId:            sp.MsgId,
	}

	request.SendRES(p)
}

//默认
type PUBRECRouter struct {
	*server.BaseRouter
}

/*
业务方可以解析 标识符合发布的标识符对应

返回  PUBRELProtocol
*/
func (s *PUBRECRouter) Handle(request *server.Request) {
	sp := request.GetProto().(*proto.PUBRECProtocol)

	fmt.Println("PUBREC msg_id === ", sp.MsgId)

	// 需要返回 PUBRELProtocol
	p := &proto.PUBRELProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PUBREL,
			MsgLen:     2,
			Data:       nil,
		},
		PacketIdentifier: sp.PacketIdentifier,
		MsgId:            sp.MsgId,
	}

	request.SendRES(p)
}

//默认
type PUBCOMPRouter struct {
	*server.BaseRouter
}

/*
业务方可以解析 标识符合发布的标识符对应
*/
func (s *PUBCOMPRouter) Handle(request *server.Request) {
	sp := request.GetProto().(*proto.PUBCOMPProtocol)

	fmt.Println("PUBCOMP msg_id === ", sp.MsgId)
}
