package proto

import (
	"errors"
)

/*
链接确认协议  服务端到客户端
CONNACK = 0x20
*/
type CONNACKProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  CONNACK = 0x20  第一个字节
	//	MsgLen = 2 对于 CONNACK 报文这个值等于 2  第二个字节
	*Fixed

	// 可变报头 这个是特有的可变报头
	ConnectAcknowledgeFlags uint8 //  第三个字节  必须是0
	ConnectReturncode       uint8 // 第四个字节 返回码 根据文档定义

	PropertiesLength                 uint32            // 1-4 字节 和 msgLen 相同
	SessionExpiryInterval            uint32            // 会话过期间隔	四字节整数
	ReceiveMaximum                   uint16            //接收最大数量
	MaximumQoS                       uint8             //
	RetainAvailable                  uint8             //
	MaximumPacketSize                uint32            //最大报文长度
	AssignedClientIdentifier         string            //分配客户标识符
	TopicAliasMaximum                uint16            // 主题别名最大长度
	ReasonString                     string            // 原因字符串	UTF-8编码字符串
	UserProperty                     map[string]string // 用户属性
	WildcardSubscriptionAvailable    uint8             // 通配符订阅可用性	字节
	SubscriptionIdentifiersAvailable uint8             // 订阅标识符可用性	字节
	SharedSubscriptionAvailable      uint8             //共享订阅可用性	字节
	ServerKeepAlive                  uint16            //服务端保活时间	双字节整数
	ResponseInformation              string            // 请求信息	UTF-8编码字符串
	ServerReference                  string            // 服务端参考	UTF-8编码字符串
	AuthenticationMethod             string            // 认证方法
	AuthenticationData               string            // 认证数据
}

func NewCONNACKProtocol(code uint8) *CONNACKProtocol {

	return &CONNACKProtocol{
		Fixed: &Fixed{
			HeaderFlag: CONNACK,
			MsgLen:     2,
		},
		ConnectAcknowledgeFlags: 0,
		ConnectReturncode:       code,
		//
		PropertiesLength:                 0,
		SessionExpiryInterval:            0,
		ReceiveMaximum:                   0,
		MaximumQoS:                       0,
		RetainAvailable:                  0,
		MaximumPacketSize:                0,
		AssignedClientIdentifier:         "",
		TopicAliasMaximum:                0,
		ReasonString:                     "",
		UserProperty:                     nil,
		WildcardSubscriptionAvailable:    0,
		SubscriptionIdentifiersAvailable: 0,
		SharedSubscriptionAvailable:      0,
		ServerKeepAlive:                  0,
		ResponseInformation:              "",
		ServerReference:                  "",
		AuthenticationMethod:             "",
		AuthenticationData:               "",
	}
}
func NewCONNACKProtocolF(f *Fixed) *CONNACKProtocol {

	return &CONNACKProtocol{
		Fixed:                   f,
		ConnectAcknowledgeFlags: 0,
		ConnectReturncode:       0,
		//
		PropertiesLength:                 0,
		SessionExpiryInterval:            0,
		ReceiveMaximum:                   0,
		MaximumQoS:                       0,
		RetainAvailable:                  0,
		MaximumPacketSize:                0,
		AssignedClientIdentifier:         "",
		TopicAliasMaximum:                0,
		ReasonString:                     "",
		UserProperty:                     nil,
		WildcardSubscriptionAvailable:    0,
		SubscriptionIdentifiersAvailable: 0,
		SharedSubscriptionAvailable:      0,
		ServerKeepAlive:                  0,
		ResponseInformation:              "",
		ServerReference:                  "",
		AuthenticationMethod:             "",
		AuthenticationData:               "",
	}
}

//func (s *CONNACKProtocol) Pack() ([]byte, error) {
//
//	// 2 打包数据 必须按照以下顺序写入
//	dataBuff := bytes.NewBuffer([]byte{})
//
//	// 2-1写入 消息类型 占据1个字节
//	if err := binary.Write(dataBuff, binary.LittleEndian, s.HeaderFlag); err != nil {
//		return []byte{}, errors.New("SendCONNACK 写入HeaderFlag 错误" + err.Error())
//	}
//	// 2-2在写入 消息长度 占据1个字节  必须转化长度 不然字节多了
//	if err := binary.Write(dataBuff, binary.LittleEndian, uint8(s.MsgLen)); err != nil {
//		return []byte{}, errors.New("SendCONNACK 写入MsgLen 错误" + err.Error())
//	}
//	// 2-3  写入 ConnectAcknowledgeFlags  占据1个字节
//	if err := binary.Write(dataBuff, binary.LittleEndian, s.ConnectAcknowledgeFlags); err != nil {
//		return []byte{}, errors.New("SendCONNACK 写入ConnectAcknowledgeFlags 错误" + err.Error())
//	}
//
//	// 2-4 写入 ConnectReturncode 一个字节
//	if err := binary.Write(dataBuff, binary.LittleEndian, s.ConnectReturncode); err != nil {
//		return []byte{}, errors.New("SendCONNACK 写入ConnectReturncode 错误" + err.Error())
//	}
//
//	return dataBuff.Bytes(), nil
//
//}
func (s *CONNACKProtocol) Pack() ([]byte, error) {
	// 固定报头
	by := make([]byte, 1, 4) // 至少4个字节
	by[0] = s.GetHeaderFlag()

	by = append(by, s.msgLenCode(s.GetMsgLen())...)

	// 可变报头
	by = append(by, s.ConnectAcknowledgeFlags, s.ConnectReturncode)

	return by, nil

}

func (s *CONNACKProtocol) UnPack() error {
	// 剩余长度至少 2
	if s.Fixed.MsgLen < 2 {
		return errors.New("报文长度错误")
	}

	s.ConnectAcknowledgeFlags = s.Fixed.Data[0]
	s.ConnectReturncode = s.Fixed.Data[1]

	if s.ConnectReturncode != Success {
		// 结束链接
		return errors.New("链接响应码错误 ")
	}
	return nil
}
