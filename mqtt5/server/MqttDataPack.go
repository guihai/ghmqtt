package server

import (
	"encoding/binary"
	"errors"
	"hshiye/tvl2mqtt/mqtt5/proto"
	"io"
)

/*
先解包 固定报头，
然后根据固定报头获取协议类型，再分别执行不同协议的解包方法
*/
type MqttDataPack struct {
}

func newMqttDataPack() *MqttDataPack {
	return &MqttDataPack{}
}

/*
打包， CONNACK 协议
返回打包后的字节
*/
func (s *MqttDataPack) packCONNACK(returncode uint8) ([]byte, error) {

	// 1，创建协议
	p := proto.NewCONNACKProtocol(returncode)

	return p.Pack()

}

/*
解包固定 报头
1,创建结构体
2，拆解数据
*/
func (s *MqttDataPack) unPackFixed(conn *Conn) (*proto.Fixed, error) {

	/*
		先获取第一个字节，这是协议类型
	*/
	HeaderFlag := make([]byte, 1)

	tcpCon := conn.getTcpConn()

	if _, err := io.ReadFull(tcpCon, HeaderFlag); err != nil {
		return nil, errors.New("获取HeaderFlag失败" + err.Error())
	}

	/*
			获取剩余长度，1到4个字节
			先获取一个字节，如果是大于127，然后继续读取
		1 个字节时，从 0(0x00)到 127(0x7f)    <= 0x7f 1个字节
		2 个字节时，从 128(0x80,0x01)到 16383(0Xff,0x7f)   在判断第二个字节 <= 0x7f  就不继续读取了
		3 个字节时，从 16384(0x80,0x80,0x01)到 2097151(0xFF,0xFF,0x7F)  判断第三个字节 <=0x7f 就不继续读取了
		4 个字节时，从 2097152(0x80,0x80,0x80,0x01)到 268435455(0xFF,0xFF,0xFF,0x7F)
	*/

	var msgLen = make([]uint8, 4)

	oneLen := make([]byte, 1)
	if _, err := io.ReadFull(tcpCon, oneLen); err != nil {
		return nil, errors.New("获取MsgLen[0]失败" + err.Error())
	}
	// 首先获取第一个
	msgLen[0] = oneLen[0]

	if oneLen[0] > 0x7f {
		//多个字节 需要获取下一个字节
		// 读取第二个字节,还使用同一个变量
		if _, err := io.ReadFull(tcpCon, oneLen); err != nil {
			return nil, errors.New("获取MsgLen[1]失败" + err.Error())
		}
		// 获取第二个字节
		msgLen[1] = oneLen[0]

		if oneLen[0] > 0x7f {
			// 需要获取第三个字节
			if _, err := io.ReadFull(tcpCon, oneLen); err != nil {
				return nil, errors.New("获取MsgLen[2]失败" + err.Error())
			}
			// 获取第三个字节
			msgLen[2] = oneLen[0]

			if oneLen[0] > 0x7f {
				// 需要获取第四个字节
				if _, err := io.ReadFull(tcpCon, oneLen); err != nil {
					return nil, errors.New("获取MsgLen[3]失败" + err.Error())
				}
				// 获取第四个字节
				msgLen[3] = oneLen[0]
			}

		}

	}

	// 长度解码
	dataLen := s.msgLenEnCode(msgLen)

	// dataLen 可以为空
	if HeaderFlag[0] < 1 {
		// 数据获取错误
		return nil, errors.New("获取数据错误")
	}

	data := make([]byte, dataLen)
	if dataLen > 0 {
		// 获取剩余字节数据
		if _, err := io.ReadFull(conn.getTcpConn(), data); err != nil {
			return nil, errors.New("获取剩余 字节数据失败 " + err.Error())
		}
	}

	// 以上完成长度获取
	return &proto.Fixed{
		HeaderFlag: HeaderFlag[0],
		MsgLen:     dataLen,
		// 需要剩余字节，后续解析协议使用
		Data: data,
	}, nil
}

/*
拆解 CONNECTProtocol 协议
1,拆解固定报头
2，拆解可变报头
3，拆解有效载荷
4,返回
*/
func (s *MqttDataPack) unPackCONNECTProtocol(conn *Conn) (*proto.CONNECTProtocol, uint8) {

	// 1，拆解固定报头
	f, err := s.unPackFixed(conn)

	if err != nil {
		return nil, proto.Malformed_Packet
	}

	// 2，使用 协议 自带的解包方法
	p := proto.NewCONNECTProtocol(f)

	p.UnPack()

	return p, p.AckCode
}

/*
固定报头 剩余长度解码算法
错误值返回 0

固定报头中剩余长度 是128进制
https://blog.csdn.net/caofengtao1314/article/details/116482822

*/
func (s *MqttDataPack) msgLenEnCode(by []byte) uint32 {

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
func (s *MqttDataPack) msgLenCode(sln uint32) []byte {

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

/*
根据固定头部解析协议类型
客户端到服务端

且不可以是 链接标志

*/
func (s *MqttDataPack) getProtoByFixed(f *proto.Fixed) (proto.ImplMqttProto, error) {

	var p proto.ImplMqttProto
	var err error = nil

	flag := f.GetHeaderFlag()
	// 不可以是连接协议
	if int(flag) == proto.CONNECT {
		return p, errors.New("CONNECT 只能出现一次")
	}

	// 订阅消息比较特殊单独处理
	// 订阅消息最小值 48 最大值 61  才有效
	if flag >= proto.PUBLISH && flag <= proto.PUBLISHMAX {
		p = proto.NewPUBLISHProtocol(f)
		err = p.UnPack()
		return p, err
	}

	switch flag {
	case proto.PUBACK:
		// 解析 PUBACK
		p = proto.NewPUBACKProtocolF(f)
		err = p.UnPack()
	case proto.PUBREC:
		// 解析 PUBREC
		p = proto.NewPUBRECProtocolF(f)
		err = p.UnPack()
	case proto.PUBREL:
		// 解析 PUBREL
		p = proto.NewPUBRELProtocol(f)
		err = p.UnPack()

	case proto.PUBCOMP:
		//解析 PUBCOMP
		p = proto.NewPUBCOMPProtocolF(f)
		err = p.UnPack()

	case proto.SUBSCRIBE:
		// 解析订阅协议
		p = proto.NewSUBSCRIBEProtocol(f)
		err = p.UnPack()
	case proto.UNSUBSCRIBE:
		// 解析 取消订阅协议
		p = proto.NewUNSUBSCRIBEProtocol(f)
		err = p.UnPack()

	case proto.PINGREQ:
		//PINGREQ 心跳请求协议
		p = proto.NewPINGREQProtocol(f)

	case proto.DISCONNECT:
		// 断开链接 协议
		p = proto.NewDISCONNECTProtocol(f)
		// 拆包
		err = p.UnPack()

	case proto.AUTH:
		// 认证协议
		p = proto.NewAUTHProtocolF(f)
		// 拆包
		err = p.UnPack()

	default:
		return p, errors.New("没有匹配到协议")
	}

	return p, err
}

func (s *MqttDataPack) int16ToByBig(ua uint16) []byte {
	var be = make([]byte, 2) // 大端
	// 大端写入  前面的16进制在前 ===》 大端写出
	binary.BigEndian.PutUint16(be, ua)
	return be
}
