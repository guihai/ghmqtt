package proto

import (
	"bytes"
	"encoding/binary"
)

/*
客户端断开链接协议
DISCONNECT  = 0xE0 // == 224    1110 0000         C=>S
*/
type DISCONNECTProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  DISCONNECT  = 0xE0 // == 224    1110 0000         C=>S
	//	MsgLen 变长字节
	*Fixed

	// 断开原因
	ReasonCode uint8 // 断开原因码

	PropertiesLength      uint32            // 1-4 字节 和 msgLen 相同
	SessionExpiryInterval uint32            // 会话过期间隔	四字节整数
	ReasonString          string            // 原因字符串	UTF-8编码字符串
	UserProperty          map[string]string // 用户属性	字符串
	ServerReference       string            // 服务端参考	UTF-8编码字符串
}

func NewDISCONNECTProtocol(f *Fixed) *DISCONNECTProtocol {
	return &DISCONNECTProtocol{
		Fixed:                 f,
		ReasonCode:            0,
		PropertiesLength:      0,
		SessionExpiryInterval: 0,
		ReasonString:          "",
		UserProperty:          nil,
		ServerReference:       "",
	}
}

func (s *DISCONNECTProtocol) UnPack() error {

	if s.Fixed.MsgLen < 1 {
		s.ReasonCode = 0x0
		return nil
	}
	s.ReasonCode = s.Fixed.Data[0]

	if s.Fixed.MsgLen < 2 {
		return nil
	}
	// 属性
	by := s.Fixed.Data[1:]
	var idx uint32 = 1 // 至少是1,最大是4
	s.PropertiesLength, idx = s.by2Len32AndIndex(by)

	if s.PropertiesLength > 0 {
		// 有属性值，需要处理
		temp := by[idx:(idx + s.PropertiesLength)]

		var tinx uint32 = 0

		// 移动坐标  最后一位 是 len-1
		for tinx < s.PropertiesLength {

			switch temp[tinx] {
			case SessionEI:
				// 四字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+4]),
					binary.BigEndian, &s.SessionExpiryInterval)
				tinx = tinx + 1 + 4

			case ServerRef:
				//先获取长度
				var ctLen uint16 = 0
				//双字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
					binary.BigEndian, &ctLen)
				tinx = tinx + 1 + 2

				s.ServerReference = string(temp[tinx : tinx+uint32(ctLen)])

				tinx += uint32(ctLen)

			case ReasonString:
				//先获取长度
				var ctLen uint16 = 0
				//双字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
					binary.BigEndian, &ctLen)
				tinx = tinx + 1 + 2

				s.ReasonString = string(temp[tinx : tinx+uint32(ctLen)])

				tinx += uint32(ctLen)

			case UserProperty:
				// 用户属性
				// len 两个字节
				var keyLen uint16
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
					binary.BigEndian, &keyLen)

				tinx += 1 + 2
				key := string(temp[tinx : tinx+uint32(keyLen)])

				tinx += uint32(keyLen)

				var valLen uint16
				binary.Read(bytes.NewBuffer(temp[tinx:tinx+2]),
					binary.BigEndian, &valLen)

				tinx += 2

				val := string(temp[tinx : tinx+uint32(valLen)])

				tinx += uint32(valLen)

				s.UserProperty[key] = val

			default:
				// 匹配不到 结束循环
				break

			}

		}

	}

	return nil
}
