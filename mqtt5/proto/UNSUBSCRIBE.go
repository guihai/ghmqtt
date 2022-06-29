package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

/*
取消订阅协议
UNSUBSCRIBE = 0xA2 // == 162        1010 0010      C=>S
*/
type UNSUBSCRIBEProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  UNSUBSCRIBE = 0xA2 // == 162        1010 0010      C=>S
	//	MsgLen = 1-4
	*Fixed

	// 可变报头
	PacketIdentifier [2]byte // 两个字节  订阅确认还要使用相同的可变报头
	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储
	// 属性
	PropertiesLength uint32            // 1-4 字节 和 msgLen 相同
	UserProperty     map[string]string // 用户属性	字符串

	// 有效载荷   没有 QoS 所以至少是3个字节
	TopicFilterList []*TopicFilter

	// AckCode 生成对应响应使用的AckCode
	AckCode uint8
}

func NewUNSUBSCRIBEProtocol(f *Fixed) *UNSUBSCRIBEProtocol {
	return &UNSUBSCRIBEProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{},
		MsgId:            0,
		PropertiesLength: 0,
		UserProperty:     make(map[string]string),
		TopicFilterList:  make([]*TopicFilter, 0),

		//
		AckCode: Success, // 默认成功
	}
}

func (s *UNSUBSCRIBEProtocol) GetAckCode() uint8 {
	return s.AckCode
}

/*
解包 订阅主题协议
1,创造协议
2，根据 固定报头 解包
*/
func (s *UNSUBSCRIBEProtocol) UnPack() error {
	// 2 + 1 + 主题 至少 3个字节  = 6
	if s.Fixed.MsgLen < 6 {
		s.AckCode = Malformed_Packet
		return errors.New("协议数据长度错误")
	}

	s.PacketIdentifier = [2]byte{s.Fixed.Data[0], s.Fixed.Data[1]}
	// 计算标识符id  大端编码 字节写入数字
	binary.Read(bytes.NewBuffer(s.PacketIdentifier[:]),
		binary.BigEndian, &s.MsgId)

	daBy := s.Fixed.Data[2:]
	// 拆解属性
	var idx uint32 = 1
	s.PropertiesLength, idx = s.by2Len32AndIndex(daBy)
	daBy = daBy[idx:]
	if s.PropertiesLength > 0 {

		daBy = daBy[s.PropertiesLength:]
	}
	// 2 拆解有效载荷
	//一个主题 至少 4个字节 长度2，字符1，Qos1
	for len(daBy) >= 3 {
		tf := &TopicFilter{
			Identifier: 0,
			FilterName: "",
			Options:    0,
		}

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(daBy[:2]),
			binary.BigEndian, &tf.Identifier)

		tf.FilterName = string(daBy[2:(2 + tf.Identifier)]) // 从 索引[2] 取对应长度值

		s.TopicFilterList = append(s.TopicFilterList, tf)

		// 字节截取
		daBy = daBy[2+tf.Identifier:]
	}

	return nil
}
