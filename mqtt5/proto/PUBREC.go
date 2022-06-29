package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

/*
发布收到 协议
Qos = 2 的publish 需要 这个响应
发送的 PUBLISH 报文必须包含报文标识符且 QoS 等于 2，DUP 等于 0
PUBREC      = 0x50 // == 80   0101 0000         C<=>S
*/
type PUBRECProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PUBREC      = 0x50 // == 80   0101 0000         C<=>S
	// MsgLen = 2
	*Fixed

	// 可变报头
	// 等于需要确认的 PUBLISH 的  PacketIdentifier
	PacketIdentifier [2]byte // 两个字节  第一个标识 MSB   和 第二个标识LSB
	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

	ReasonCode       uint8  // 使用响应码
	PropertiesLength uint32 // 属性长度
}

func NewPUBRECProtocol(pid [2]byte, rec uint8) *PUBRECProtocol {
	p := &PUBRECProtocol{
		Fixed: &Fixed{
			HeaderFlag: PUBREC,
			MsgLen:     4, // 默认4
			Data:       nil,
		},
		// 标识符 和 publish 一样
		PacketIdentifier: pid,
		ReasonCode:       rec,
		PropertiesLength: 0,
	}

	return p
}
func NewPUBRECProtocolF(f *Fixed) *PUBRECProtocol {
	p := &PUBRECProtocol{
		Fixed: f,
		// 标识符 和 publish 一样
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]},
		ReasonCode:       0,
		PropertiesLength: 0,
	}

	return p
}

func (s *PUBRECProtocol) Pack() ([]byte, error) {

	// 至少6个字节
	by := make([]byte, 1, 6)
	// 固定报头
	by[0] = s.GetHeaderFlag()

	by = append(by, s.msgLenCode(s.GetMsgLen())...)

	// 可变报头
	by = append(by, s.PacketIdentifier[0], s.PacketIdentifier[1])

	//原因码
	by = append(by, s.ReasonCode)

	// todo 属性默认0
	by = append(by, s.msgLenCode(s.PropertiesLength)...)

	return by, nil

}

func (s *PUBRECProtocol) UnPack() error {
	// 剩余长度至少 3
	if s.Fixed.MsgLen < 3 {
		//s.AckCode = Malformed_Packet
		return errors.New("报文长度错误")
	}
	binary.Read(bytes.NewBuffer(s.PacketIdentifier[:]),
		binary.BigEndian, &s.MsgId)

	s.ReasonCode = s.Fixed.Data[2]

	// 属性长度
	if s.Fixed.MsgLen > 3 {
		// 获取属性
		by := s.Fixed.Data[3:]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// todo 获取属性
		}
	}
	return nil
}
