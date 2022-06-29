package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

/*
订阅协议
SUBSCRIBE   = 0x82 //  == 130        1000 0010     C=>S
*/

type SUBSCRIBEProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  SUBSCRIBE   = 0x82 //  == 130        1000 0010     C=>S
	//	MsgLen = 1-4   ， PacketIdentifier 两个字节 有效载荷多个
	*Fixed

	// 可变报头
	PacketIdentifier [2]byte // 两个字节
	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储
	// 属性
	PropertiesLength       uint32            // 1-4 字节 和 msgLen 相同
	SubscriptionIdentifier uint32            // 变长字节整数
	UserProperty           map[string]string // 用户属性	字符串
	// 有效载荷
	TopicFilterList []*TopicFilter

	// AckCode 生成对应响应使用的AckCode
	AckCode uint8
}

// SUBSCRIBEProtocol 使用
type TopicFilter struct {
	Identifier uint16 // 两个字节
	FilterName string // 根据Identifier 长度获取
	//QoS        uint8  // 本组最后一个字节 只能 从常量  QoS0,1,2中取值
	// todo 包含  Retain Handling 、RAP、NL、Qos
	Options uint8 // 本组最后一个字节 只能 从常量  QoS0,1,2中取值
}

func NewSUBSCRIBEProtocol(f *Fixed) *SUBSCRIBEProtocol {
	return &SUBSCRIBEProtocol{
		Fixed:                  f,
		PacketIdentifier:       [2]byte{},
		MsgId:                  0,
		PropertiesLength:       0,
		SubscriptionIdentifier: 0,
		UserProperty:           make(map[string]string),
		TopicFilterList:        make([]*TopicFilter, 0),

		//
		AckCode: Success, // 默认成功
	}
}

func (s *SUBSCRIBEProtocol) GetAckCode() uint8 {
	return s.AckCode
}

/*
解包 订阅主题协议
1,创造协议
2，根据 固定报头 解包
*/
func (s *SUBSCRIBEProtocol) UnPack() error {
	// 2 + 1 + 主题 至少 4个字节  = 7
	if s.Fixed.MsgLen < 7 {
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
	for len(daBy) >= 4 {
		tf := &TopicFilter{
			Identifier: 0,
			FilterName: "",
			Options:    0,
		}

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(daBy[:2]),
			binary.BigEndian, &tf.Identifier)

		tf.FilterName = string(daBy[2:(2 + tf.Identifier)]) // 从 索引[2] 取对应长度值
		tf.Options = daBy[2+tf.Identifier]                  // Qos

		s.TopicFilterList = append(s.TopicFilterList, tf)

		// 字节截取
		daBy = daBy[2+tf.Identifier+1:]
	}

	return nil
}
