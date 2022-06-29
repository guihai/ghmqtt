package client

import (
	"hshiye/tvl2mqtt/mqtt5/proto"
	"net"
)

type Message struct {
	// 发送或者收到的协议
	proto proto.ImplMqttProto
	//
	dp *DataPack
	//
	ofConn net.Conn
}

func NewMessage(con net.Conn) *Message {

	return &Message{
		proto:  nil,
		dp:     newDataPack(),
		ofConn: con,
	}
}

func (s *Message) getMqttProto() error {

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
