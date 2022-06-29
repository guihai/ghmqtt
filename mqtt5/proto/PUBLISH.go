package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

/*
发布协议
PUBLISH     = 0x30 // == 48   0011 xxxx   大于等于48 小于 64   C<=>S
*/
type PUBLISHProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PUBLISH     = 0x30 // == 48   0011 xxxx   大于等于48 小于 64   C<=>S
	*Fixed

	// 可变报头
	TopicNameLength  uint16 // 两个字节  大端解码返回数值
	TopicName        string
	PacketIdentifier [2]byte // 两个字节  第一个标识 MSB   和 第二个标识LSB
	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

	// 属性
	PropertiesLength       uint32            // 1-4 字节 和 msgLen 相同
	PayloadFormatIndicator uint8             // 载荷格式说明	字节
	MessageExpiryInterval  uint32            //消息过期时间	四字节整数
	TopicAlias             uint16            //主题别名	双字节整数
	ResponseTopic          string            //响应主题	UTF-8编码字符串
	CorrelationData        string            // 对比数据
	UserProperty           map[string]string // 用户属性
	SubscriptionIdentifier uint32            // 定义标识符	变长字节整数
	ContentType            string

	// 有效载荷  剩余字节都是主题内容
	Payload []byte

	// 解析Qos 用于处理
	Qos uint8

	// 解析 Retain 用于处理
	Retain bool // 0 或者1

	// AckCode 生成对应响应使用的AckCode
	AckCode uint8
}

func NewPUBLISHProtocol(f *Fixed) *PUBLISHProtocol {
	return &PUBLISHProtocol{
		Fixed:                  f,
		TopicNameLength:        0,
		TopicName:              "",
		PacketIdentifier:       [2]byte{},
		MsgId:                  0,
		PropertiesLength:       0,
		PayloadFormatIndicator: 0,
		MessageExpiryInterval:  0,
		TopicAlias:             0,
		ResponseTopic:          "",
		CorrelationData:        "",
		UserProperty:           nil,
		SubscriptionIdentifier: 0,
		ContentType:            "",
		Payload:                nil,
		Qos:                    QoS0,    // 默认0
		Retain:                 false,   //默认
		AckCode:                Success, // 默认成功
	}
}

func (s *PUBLISHProtocol) GetAckCode() uint8 {
	return s.AckCode
}

func (s *PUBLISHProtocol) UnPack() error {

	// 至少7个字节
	if s.Fixed.MsgLen < 5 {
		s.AckCode = Malformed_Packet
		return errors.New("报文长度错误")
	}

	binary.Read(bytes.NewBuffer(s.Fixed.Data[:2]),
		binary.BigEndian, &s.TopicNameLength)

	s.TopicName = string(s.Fixed.Data[2:(2 + s.TopicNameLength)])

	switch s.Fixed.HeaderFlag {
	case PUBLISH:
		// 48 没有 PacketIdentifier
		// 获取属性
		by := s.Fixed.Data[(2 + s.TopicNameLength):]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// 获取属性
			by = by[s.PropertiesLength:]
		}

		s.Payload = by
		s.Qos = QoS0

	case PUBLISH31:
		// 49 没有 PacketIdentifier 有保留位
		by := s.Fixed.Data[(2 + s.TopicNameLength):]
		// 获取属性
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// 获取属性
			by = by[s.PropertiesLength:]
		}

		s.Payload = by
		s.Qos = QoS0
		s.Retain = true

	case PUBLISH32:
		// 50
		s.PacketIdentifier = [2]byte{s.Fixed.Data[2+s.TopicNameLength],
			s.Fixed.Data[2+s.TopicNameLength+1]}
		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(s.PacketIdentifier[:]),
			binary.BigEndian, &s.MsgId)
		// 获取属性
		by := s.Fixed.Data[(2 + s.TopicNameLength + 2):]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// 获取属性
			by = by[s.PropertiesLength:]
		}

		s.Payload = by

		//p.Payload = f.Data[(2 + p.TopicNameLength + 1 + 1):]
		s.Qos = QoS1

	case PUBLISH33:
		// 51
		s.PacketIdentifier = [2]byte{s.Fixed.Data[2+s.TopicNameLength],
			s.Fixed.Data[2+s.TopicNameLength+1]}
		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(s.PacketIdentifier[:]),
			binary.BigEndian, &s.MsgId)
		// 获取属性
		by := s.Fixed.Data[(2 + s.TopicNameLength + 2):]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// 获取属性
			by = by[s.PropertiesLength:]
		}

		s.Payload = by

		s.Qos = QoS1
		s.Retain = true

	case PUBLISH34:
		// 52
		s.PacketIdentifier = [2]byte{s.Fixed.Data[2+s.TopicNameLength],
			s.Fixed.Data[2+s.TopicNameLength+1]}
		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(s.PacketIdentifier[:]),
			binary.BigEndian, &s.MsgId)

		// 获取属性
		by := s.Fixed.Data[(2 + s.TopicNameLength + 2):]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// 获取属性
			by = by[s.PropertiesLength:]
		}
		s.Payload = by
		s.Qos = QoS2

	default:
		// 其他 Qos1,2 都要有标识符
		s.PacketIdentifier = [2]byte{s.Fixed.Data[2+s.TopicNameLength],
			s.Fixed.Data[2+s.TopicNameLength+1]}
		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(s.PacketIdentifier[:]),
			binary.BigEndian, &s.MsgId)
		// 获取属性
		by := s.Fixed.Data[(2 + s.TopicNameLength + 2):]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// 获取属性
			by = by[s.PropertiesLength:]
		}

		s.Payload = by
		// todo 暂定1
		s.Qos = QoS1
	}

	return nil
}

func (s *PUBLISHProtocol) Pack() ([]byte, error) {

	// 固定报头
	by := make([]byte, 1, 8) // 至少8个字节
	by[0] = s.GetHeaderFlag()

	by = append(by, s.msgLenCode(s.GetMsgLen())...)

	// 可变报头
	by = append(by, s.int16ToByBig(s.TopicNameLength)...)
	by = append(by, []byte(s.TopicName)...)

	// 根据 报头 ，确定是否有 标识符
	if s.Qos > QoS0 {
		// 需要 获取标识符
		by = append(by, s.int16ToByBig(s.MsgId)...)
	}

	// todo 属性默认0
	by = append(by, s.msgLenCode(s.PropertiesLength)...)
	// 有效载荷
	by = append(by, s.Payload...)

	return by, nil

}
