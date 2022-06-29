package proto

import (
	"bytes"
	"encoding/binary"
)

// 固定报头 通用型 所有协议都有固定报头，都继承此结构体，同时实现了 接口
type Fixed struct {
	// 固定报头
	HeaderFlag uint8  // 协议类型 第1个字节
	MsgLen     uint32 // 剩余长度 1-4个字节
	// 剩余字节
	Data []byte
}

func (s *Fixed) GetHeaderFlag() uint8 {
	return s.HeaderFlag
}

func (s *Fixed) GetMsgLen() uint32 {
	return s.MsgLen
}

func (s *Fixed) GetData() []byte {
	return s.Data
}

func (s *Fixed) Pack() ([]byte, error) {

	return nil, nil

}

func (s *Fixed) UnPack() error {

	return nil

}

func (s *Fixed) GetAckCode() uint8 {

	return Success

}

/*
大端 字节转 长度和名称
失败 返回 false
*/
func (s *Fixed) by2LenNameBE(by []byte) (blen uint16, name string, ok bool) {
	buf := bytes.NewBuffer(by[:2]) // 转buf

	// 字节读入数值  大端读入 两个字节 读入 uint16
	err2 := binary.Read(buf, binary.BigEndian, &blen)

	if err2 != nil || blen == 0 {
		return
	}

	// 防止溢出
	if len(by) < int(2+blen) {
		return
	}

	name = string(by[2 : 2+blen])
	ok = true

	return

}

// 变长字节长度获取
func (s *Fixed) by2Len32(by []byte) uint32 {

	var msgLen = make([]uint8, 4)
	// 首先获取第一个
	msgLen[0] = by[0]

	if msgLen[0] > 0x7f {
		//多个字节 需要获取下一个字节
		msgLen[1] = by[1]

		if msgLen[1] > 0x7f {
			// 需要获取第三个字节
			// 获取第三个字节
			msgLen[2] = by[2]

			if msgLen[2] > 0x7f {
				// 需要获取第四个字节
				// 获取第四个字节
				msgLen[3] = by[3]
			}

		}

	}

	// 长度解码
	return s.msgLenEnCode(msgLen)
}

// 获取长度，且计算用了几个字节
func (s *Fixed) by2Len32AndIndex(by []byte) (uint32, uint32) {

	var msgLen = make([]uint8, 4)
	var count uint32 = 1
	// 首先获取第一个
	msgLen[0] = by[0]

	if msgLen[0] > 0x7f {
		//多个字节 需要获取下一个字节
		msgLen[1] = by[1]
		count += 1

		if msgLen[1] > 0x7f {
			// 需要获取第三个字节
			// 获取第三个字节
			msgLen[2] = by[2]

			count += 1

			if msgLen[2] > 0x7f {
				// 需要获取第四个字节
				// 获取第四个字节
				msgLen[3] = by[3]
				count += 1
			}

		}

	}

	// 长度解码
	return s.msgLenEnCode(msgLen), count
}

// 字节转长度
func (s *Fixed) msgLenEnCode(by []byte) uint32 {

	multiplier := 1
	sln := 0

	if len(by) < 1 || len(by) > 4 {
		// 1- 4 个字节
		return 0
	}
	for _, b := range by {

		bb := int(b)

		sln += (bb & 127) * multiplier
		multiplier *= 128

		if multiplier > 128*128*128 {
			break
		}

	}

	// 错误值返回 0
	return uint32(sln)
}

/*
固定报头 剩余长度编码算法
错误值 返回 空数组
剩余长度编码  128进制
十进制 长度 / 128

*/
func (s *Fixed) msgLenCode(sln uint32) []byte {

	// 至少有一位
	var by []byte
	if sln > 268435455 {

		// 超过最大值了  直接返回空
		return by
	}

	// 10进制除以128
	for {
		// 计算
		x := sln / 128
		y := uint8(sln % 128)

		if x > 0 {
			// 首先获取字节    按位或运算符"|"是双目运算符
			by = append(by, y|128)
			// x 大于0
			// 新的长度等于 x
			sln = x
		} else {
			by = append(by, y)
			// 不大于0 就结束
			return by
		}

	}

}

func (s *Fixed) int16ToByBig(ua uint16) []byte {
	var be = make([]byte, 2) // 大端
	// 大端写入  前面的16进制在前 ===》 大端写出
	binary.BigEndian.PutUint16(be, ua)
	return be
}

func (s *Fixed) unPackPropertyLength(by []byte) (uint32, []byte) {

	pLength, idx := s.by2Len32AndIndex(by)
	return pLength, by[idx:]
}

// 可变报头，部分控制报文包含 通用型
type Variable struct {
	ProtoNameLen uint16 // 两个字节
	ProtoName    string // 根据ProtoNameLen 长度获取的数据
	Version      uint8  // 一个字节 协议级别 版本号 3.11 = 0x04
	ConnectFlag  uint8  // 一个字节  链接标志 可以控制有效载荷
	KeepAlive    uint8  // 两个字节 保持连接 Keep Alive MSB 和 保持连接 Keep Alive LSB
}

// 遗嘱消息结构体
type Will struct {
	WillTopic   string
	WillMessage string
	WillRetain  bool
	WillQos     uint8

	PayloadFormatIndicator uint8
	MessageExpiryInterval  uint32
	ContentType            string
	ResponseTopic          string
	CorrelationData        []byte
	WillDelayInterval      uint32
}
