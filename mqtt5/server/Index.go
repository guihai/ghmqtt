package server

import (
	"github.com/guihai/ghmqtt/mqtt5/proto"
	"github.com/guihai/ghmqtt/mqtt5/server/types"
	"github.com/guihai/ghmqtt/utils"
	"math/rand"
	"time"
)

/*
入口文件，对外暴露方法
*/

type GHapi struct {
	// 基础服务
	server *Server
}

func NewGHapi() *GHapi {
	return &GHapi{
		server: newServer(),
	}
}

// 启动服务
func (s *GHapi) Run() {
	s.server.run()
}

// 停止服务
func (s *GHapi) Stop() {
	s.server.start()
}

/*
注册路由
*/
func (s *GHapi) AddRouter(i uint8, router ImplBaseRouter) {
	s.server.routerMer.addRouter(i, router)
}

/*
注册链接验证
*/
func (s *GHapi) SetConnectVerify(cvf ConnectVerifyFUNC) {
	s.server.routerMer.setConnectVerify(cvf)
}

// 对http服务和管理者客户端暴露的接口,返回值都是结构体
func (s *GHapi) ServerInfo() *types.Response {
	back := types.NewResponse()

	back.Code = utils.RECODE_OK
	back.Msg = utils.MsgText(utils.RECODE_OK)

	back.Data = map[string]interface{}{
		"IP":   s.server.ip,
		"Port": s.server.port,
		// 获取链接对象个数
		"LenConn": s.server.connMer.getLen(),
	}

	return back
}

/*
管理服务端 直接发布信息，不能发布保留信息(保留信息使用接口  SetRetainMsg)
创造协议 管理方发布的信息直接发布，不用进入 topicmananger 的协程池
*/
func (s *GHapi) SendPublish(msg *types.PublishMsg) *types.Response {
	back := types.NewResponse()

	//1 创建协议
	// todo 48 测试
	p := &proto.PUBLISHProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PUBLISH,
			MsgLen:     0,
			Data:       nil,
		},
		TopicNameLength:  uint16(len(msg.TopicName)),
		TopicName:        msg.TopicName,
		PacketIdentifier: [2]byte{},
		Payload:          []byte(msg.TopicMsg),
		Qos:              msg.Qos,
		Retain:           msg.Retain,
		MsgId:            0,
	}

	// 完善协议
	if p.Qos > proto.QoS0 {
		// 需要标识符
		rand.Seed(time.Now().Unix())
		p.MsgId = uint16(rand.Intn(50000))
		p.MsgLen = 2

		if p.Qos == proto.QoS2 {
			p.HeaderFlag = proto.PUBLISH34

		} else {
			// Qos1 处理
			p.HeaderFlag = proto.PUBLISH32
		}
	}

	p.MsgLen = p.MsgLen + 1 + 1 + uint32(p.TopicNameLength) + uint32(len(p.Payload))

	// 打包
	by, err := p.Pack()

	if err != nil {
		back.Code = 5001
		back.Msg = err.Error()
	} else {

		// 3 多个匹配条件发送
		s.server.topicMer.tm.matchSend(p.TopicName, by)
		back.Code = utils.RECODE_OK
		back.Msg = utils.MsgText(utils.RECODE_OK)
	}

	return back
}

// 设置保留消息
func (s *GHapi) SetRetainMsg(msg *types.PublishMsg) *types.Response {
	back := types.NewResponse()

	s.server.topicMer.setRetainMsg(msg.TopicName, []byte(msg.TopicMsg))

	back.Code = utils.RECODE_OK
	back.Msg = utils.MsgText(utils.RECODE_OK)
	return back
}

// 获取链接列表
func (s *GHapi) GetConnList() *types.Response {
	back := types.NewResponse()

	list := s.server.connMer.getConnList()

	back.Code = utils.RECODE_OK
	back.Msg = utils.MsgText(utils.RECODE_OK)
	back.Data = list
	return back
}

// 获取订阅主题列表
func (s *GHapi) GetTopList() *types.Response {
	back := types.NewResponse()

	list := s.server.topicMer.getTopList()

	back.Code = utils.RECODE_OK
	back.Msg = utils.MsgText(utils.RECODE_OK)
	back.Data = list
	return back
}

/*
todo 发送给谁
发送验证
*/
func (s *GHapi) SendAuth() *types.Response {
	back := types.NewResponse()

	//1 创建协议
	p := proto.NewAUTHProtocol()
	// 打包
	by, err := p.Pack()

	if err != nil {
		back.Code = 5001
		back.Msg = err.Error()
	} else {

		// 3 多个匹配条件发送
		s.server.connMer.connMap["cli-gy"].sendByte(by)

		back.Code = utils.RECODE_OK
		back.Msg = utils.MsgText(utils.RECODE_OK)
	}

	return back
}
