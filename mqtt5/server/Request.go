package server

import (
	"hshiye/tvl2mqtt/mqtt5/proto"
)

/*
服务-链接-请求，请求是主要的操作入口，
*/
type Request struct {
	// 所属的链接对象
	ofConn *Conn // 封装的连接对象

	// 解析后的协议 封装请求数据成为协议对象
	proto proto.ImplMqttProto

	// 拆包工具
	dp *MqttDataPack
}

func newRequest(conn *Conn) *Request {
	return &Request{
		ofConn: conn,
		dp:     newMqttDataPack(),
	}
}

// 获取连接对象
func (s *Request) getConn() *Conn {
	return s.ofConn
}

/*
接收 第一次连接 协议
1，创建链接协议
2，解包工具进行解包
3，将有效数据返回

*/
func (s *Request) getCONNECT() (*proto.CONNECTProtocol, uint8) {

	p, code := s.dp.unPackCONNECTProtocol(s.ofConn)

	if code != proto.Success {
		return nil, code
	}

	s.proto = p

	return p, code

}

/*
创建 链接后的返回值协议
1,创建协议结构体
2,打包数据
3，发送协议
*/
func (s *Request) sendCONNACK(returncode uint8) error {

	by, err := s.dp.packCONNACK(returncode)

	if err != nil {
		return err
	}

	// 发送数据
	s.ofConn.sendByte(by)

	return nil
}

/*
获取协议返回
主要操作入口
*/
func (s *Request) getMqttProto() error {

	// 获取固定报头
	f, err := s.dp.unPackFixed(s.ofConn)

	if err != nil {
		return err
	}

	// 根据头部解析协议类型
	p, err := s.dp.getProtoByFixed(f)

	if err != nil {
		// 解析协议错误，不操作
		return err
	}

	s.proto = p
	return nil

}

///////////////////////////////////////////////////////////////////////////
// 对外暴露接口
/*
发送响应
根据协议找到路由进行执行
*/

func (s *Request) SendRES(p proto.ImplMqttProto) error {

	by, err := p.Pack()

	if err != nil {
		return err
	}
	// 发送数据
	s.ofConn.sendByte(by)

	return nil
}

// 优化后的 取消订阅方法路径
func (s *Request) UnSubTopic(top string) {
	s.ofConn.ofServer.topicMer.unSubTopic(top, s.ofConn.clientID)
}

// 优化后的 订阅方法路径
func (s *Request) SubTopic(top string) {
	s.ofConn.ofServer.topicMer.subTopic(top, s.ofConn.clientID)
}

//  优化后的 发布方法路径
//func (s *Request) MsgIn(name string, payload []byte) {
//	s.ofConn.ofServer.topicMer.MsgIn(name, payload)
//}

// 优化后的  保存保留消息
func (s *Request) SetRetainMsg(name string, payload []byte) {
	s.ofConn.ofServer.topicMer.setRetainMsg(name, payload)
}

func (s *Request) GetQos2ID(id uint16) bool {
	return s.ofConn.ofServer.topicMer.getQos2ID(id)
}

func (s *Request) SetQos2ID(id uint16) {
	s.ofConn.ofServer.topicMer.setQos2ID(id)
}

func (s *Request) RemoveQos2ID(id uint16) {
	s.ofConn.ofServer.topicMer.removeQos2ID(id)
}

func (s *Request) MsgInPool(sp *proto.PUBLISHProtocol) {
	s.ofConn.ofServer.topicMer.msgInPool(sp)
}

func (s *Request) ConnStop() {
	s.ofConn.stop()
}

func (s *Request) GetProto() proto.ImplMqttProto {
	return s.proto
}
func (s *Request) GetConnClientID() string {
	return s.ofConn.clientID
}
