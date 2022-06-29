package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

/*
发布完成协议  Qos2 最后响应
PUBCOMP     = 0x70 //  ==112       0111 0000       C<=>S
*/
type PUBCOMPProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PUBCOMP     = 0x70 //  ==112       0111 0000       C<=>S
	// MsgLen
	*Fixed

	// 可变报头
	// 等于需要确认的 PUBLISH 的  PacketIdentifier
	PacketIdentifier [2]byte // 两个字节
	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

	ReasonCode       uint8  // 使用响应码
	PropertiesLength uint32 // 属性长度

	// AckCode 生成对应响应使用的AckCode
	AckCode uint8
}

func NewPUBCOMPProtocol(pid [2]byte, rec uint8) *PUBCOMPProtocol {
	return &PUBCOMPProtocol{
		Fixed: &Fixed{
			HeaderFlag: PUBCOMP,
			MsgLen:     4, // 默认4
			Data:       nil,
		},
		PacketIdentifier: pid,
		ReasonCode:       rec,
		PropertiesLength: 0, // 默认0
		// 默认成功
		AckCode: Success,
	}
}

func NewPUBCOMPProtocolF(f *Fixed) *PUBCOMPProtocol {
	return &PUBCOMPProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]},
		ReasonCode:       0,
		PropertiesLength: 0, // 默认0
		// 默认成功
		AckCode: Success,
	}
}

func (s *PUBCOMPProtocol) Pack() ([]byte, error) {

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

func (s *PUBCOMPProtocol) UnPack() error {
	// 剩余长度至少 3
	if s.Fixed.MsgLen < 3 {
		s.AckCode = Malformed_Packet
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
