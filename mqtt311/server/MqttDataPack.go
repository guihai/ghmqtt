package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/guihai/ghmqtt/mqtt311/proto"
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
获取proto 进行打包分发
*/
func (s *MqttDataPack) packProto(p proto.ImplMqttProto) ([]byte, error) {

	// 根据 固定头部协议类型 选择打包方法
	var by []byte
	var err error

	switch p.GetHeaderFlag() {
	case proto.PINGRESP:
		// 心跳响应
		by, err = s.packPINGRESP(p)
	case proto.SUBACK:
		// 订阅响应
		by, err = s.packSUBACK(p)
	case proto.UNSUBACK:
		// 取消订阅响应
		by, err = s.packUNSUBACK(p)
	case proto.PUBLISH:
		// 发布消息协议
		by, err = s.packPUBLISH(p)

	case proto.PUBACK:
		// 发布消息 Qos1 dup 0 需要返回的响应
		by, err = s.packPUBACK(p)

	case proto.PUBREC:
		// Qos2 第一个响应
		by, err = s.packPUBREC(p)

	case proto.PUBREL:

		by, err = s.packPUBREL(p)

	case proto.PUBCOMP:
		// Qos2 第二个响应，最后一个
		by, err = s.packPUBCOMP(p)

	default:
		err = errors.New("没有找到响应 的协议类型")

	}

	return by, err
}

/*
打包， CONNACK 协议
返回打包后的字节
*/
func (s *MqttDataPack) packCONNACK(returncode uint8) ([]byte, error) {

	// 1，创建协议
	p := &proto.CONNACKProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.CONNACK,
			MsgLen:     2,
		},
		ConnectAcknowledgeFlags: 0,
		ConnectReturncode:       returncode,
	}

	// 2 打包数据 必须按照以下顺序写入
	dataBuff := bytes.NewBuffer([]byte{})

	// 2-1写入 消息类型 占据1个字节
	if err := binary.Write(dataBuff, binary.LittleEndian, p.HeaderFlag); err != nil {
		return []byte{}, errors.New("SendCONNACK 写入HeaderFlag 错误" + err.Error())
	}
	// 2-2在写入 消息长度 占据1个字节  必须转化长度 不然字节多了
	if err := binary.Write(dataBuff, binary.LittleEndian, uint8(p.MsgLen)); err != nil {
		return []byte{}, errors.New("SendCONNACK 写入MsgLen 错误" + err.Error())
	}
	// 2-3  写入 ConnectAcknowledgeFlags  占据1个字节
	if err := binary.Write(dataBuff, binary.LittleEndian, p.ConnectAcknowledgeFlags); err != nil {
		return []byte{}, errors.New("SendCONNACK 写入ConnectAcknowledgeFlags 错误" + err.Error())
	}

	// 2-4 写入 ConnectReturncode 一个字节
	if err := binary.Write(dataBuff, binary.LittleEndian, p.ConnectReturncode); err != nil {
		return []byte{}, errors.New("SendCONNACK 写入ConnectReturncode 错误" + err.Error())
	}

	return dataBuff.Bytes(), nil

}

/*
打包 心跳响应协议
*/
func (s *MqttDataPack) packPINGRESP(p proto.ImplMqttProto) ([]byte, error) {

	// 打包数据 必须按照以下顺序写入
	dataBuff := bytes.NewBuffer([]byte{})

	// 2-1写入 消息类型 占据1个字节
	if err := binary.Write(dataBuff, binary.LittleEndian, p.GetHeaderFlag()); err != nil {
		return []byte{}, errors.New("写入HeaderFlag 错误" + err.Error())
	}
	// 2-2在写入 消息长度 占据1个字节  必须转化长度 不然字节多了
	if err := binary.Write(dataBuff, binary.LittleEndian, uint8(p.GetMsgLen())); err != nil {
		return []byte{}, errors.New("写入MsgLen 错误" + err.Error())
	}

	return dataBuff.Bytes(), nil
}

/*
打包固定头部
打包可变头部
打包有效载荷
*/
func (s *MqttDataPack) packSUBACK(p proto.ImplMqttProto) ([]byte, error) {

	p1, ok := p.(*proto.SUBACKProtocol)

	if !ok {
		return nil, errors.New("协议转化错误")
	}

	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])

	by = append(by, p1.ReturnCodeList...)

	return by, nil

}

/*
取消订阅 打包
*/
func (s *MqttDataPack) packUNSUBACK(p proto.ImplMqttProto) ([]byte, error) {

	p1, ok := p.(*proto.UNSUBACKProtocol)

	if !ok {
		return nil, errors.New("协议转化字节错误")
	}

	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])

	return by, nil

}

/*
打包发布消息 协议  用户订阅主题后，发布给用户
*/
func (s *MqttDataPack) packPUBLISH(p proto.ImplMqttProto) ([]byte, error) {

	p1, ok := p.(*proto.PUBLISHProtocol)

	if !ok {
		return nil, errors.New("协议转化字节错误")
	}

	// 固定报头
	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	// 可变报头
	// 长度 =》大端编写=》字节
	//by = append(by, 0, p1.TopicNameLength)
	by = append(by, s.int16ToByBig(p1.TopicNameLength)...)
	by = append(by, []byte(p1.TopicName)...)

	// 根据 报头 ，确定是否有 标识符
	if p1.Qos > proto.QoS0 {
		// 需要 获取标识符
		//by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])
		by = append(by, s.int16ToByBig(p1.MsgId)...)
	}

	// 有效载荷
	by = append(by, p1.Payload...)

	return by, nil

}

/*
打包 发布确认协议
*/
func (s *MqttDataPack) packPUBACK(p proto.ImplMqttProto) ([]byte, error) {

	p1, ok := p.(*proto.PUBACKProtocol)

	if !ok {
		return nil, errors.New("协议转化字节错误")
	}

	// 固定报头
	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	// 可变报头
	by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])

	return by, nil

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
func (s *MqttDataPack) unPackCONNECTProtocol(conn *Conn) (*proto.CONNECTProtocol, uint8, error) {

	// 1，拆解固定报头
	f, err := s.unPackFixed(conn)
	if err != nil {
		return nil, proto.Refused_u_p_v, err
	}

	// 创造协议
	p := &proto.CONNECTProtocol{
		Fixed:          f,
		ProtoNameLen:   0,
		ProtoName:      "",
		Version:        0,
		ConnectFlag:    0,
		KeepAlive:      0,
		ClientIDLength: 0,
		ClientID:       "",

		CleanSession: true,       // 默认true
		WillFlag:     false,      // 默认false
		WillRetain:   false,      // 默认false
		WillQos:      proto.QoS0, // 默认0
	}
	// 2 拆解可变报头
	daBy := f.Data

	// 计算标识符id  大端编码 字节写入数字
	binary.Read(bytes.NewBuffer(daBy[:2]),
		binary.BigEndian, &p.ProtoNameLen)

	p.ProtoName = string(daBy[2:(2 + p.ProtoNameLen)]) // 从 索引[2] 取对应长度值
	p.Version = daBy[2+p.ProtoNameLen]                 // 版本号  4 == 版本 3.1.1
	p.ConnectFlag = daBy[2+p.ProtoNameLen+1]           // 后面的字节 是 连接标志
	// 计算标识符id  大端编码 字节写入数字
	binary.Read(bytes.NewBuffer(daBy[2+p.ProtoNameLen+1+1:2+p.ProtoNameLen+1+1+1+1]),
		binary.BigEndian, &p.KeepAlive)

	// 第一次链接必须进行基础判断
	// 第一次链接 必须进行 检测 mqtt，版本号 0x04，协议类型 0x10,必须有client,
	if p.ProtoName != "MQTT" || p.HeaderFlag != proto.CONNECT || p.Version != 0x04 {

		return nil, proto.Refused_u_p_v, errors.New("必须是 MQTT 3.1.1 版本")
	}

	// 3 拆解有效载荷
	daByp := daBy[2+p.ProtoNameLen+1+1+1+1:] // 获取有效载荷的后续数据

	if len(daByp) < 3 {
		// 第一次链接必须有 client 必须有值
		return nil, proto.Refused_i_r, errors.New("CONNECT 有效载荷不能为空")
	}

	p.ClientIDLength, p.ClientID, _ = s.by2LenNameBE(daByp)

	if p.ClientIDLength < 1 || p.ClientID == "" {
		return p, proto.Refused_i_r, errors.New("获取 ClientIDLength 错误" + err.Error())
	}

	daByp2 := daByp[2+p.ClientIDLength:] // 获取后续数据

	// 数字变成 8位二级制
	bs := fmt.Sprintf("%08b", p.ConnectFlag)

	// 根据  ConnectFlag 获取其他数据
	// 48 == 0 ，49 ==1
	if bs[7] != 48 {
		// CONNECT 控制报文的保留标志位（第 0 位）是否为 0，如果不为 0 必须断开客户端 连接
		return p, proto.Refused_u_p_v, nil
	}

	// CleanSession  1代表清除会话，不保留离线消息 ，0代表保留会话，当连接断开后，客户端和服务端 必须保存会话信息
	if bs[6] == 48 { // 0
		// 1 不用处理 0 要保留回话，在业务端处理
		p.CleanSession = false // 业务端需要处理 false
	}

	// Will Flag  遗嘱标志  遗嘱主题，遗嘱消息，当客户端断开链接时，要将客户端的遗嘱 向订阅遗嘱主题的客户端发布
	// 让其他客户端 获取某个 客户端下线
	if bs[5] == 49 { // 1 需要 在有效载荷中获取 遗嘱数据
		// 获取 遗嘱有效数据 遗嘱主题和遗嘱消息 至少6个字节
		if len(daByp2) < 6 {
			// 不存在遗嘱
			return p, proto.Refused_i_r, nil

		}
		p.WillTopicLength, p.WillTopic, _ = s.by2LenNameBE(daByp2)
		if p.WillTopicLength < 1 || p.WillTopic == "" {
			return p, proto.Refused_i_r, nil
		}

		// 获取遗嘱主题成功，获取遗嘱消息
		p.WillMessageLength, p.WillMessage, _ = s.by2LenNameBE(daByp2[2+p.WillTopicLength:])
		if p.WillMessageLength < 1 || p.WillMessage == "" {
			return p, proto.Refused_i_r, nil
		}

		p.WillFlag = true

		// 需要处理遗嘱相关标识
		if bs[2] == 49 {
			p.WillRetain = true

		}

		// 01
		if bs[3] == 48 && bs[4] == 49 {
			p.WillQos = proto.QoS1
		}

		// 10
		if bs[3] == 49 && bs[4] == 48 {
			p.WillQos = proto.QoS2
		}

		// 以上获取遗嘱完成  修改剩余 字节的长度  必须修改
		daByp2 = daByp2[2+p.ClientIDLength+2+p.WillMessageLength+1:]

	}

	// User Name Flag
	if bs[0] == 49 {
		if len(daByp2) < 3 {
			// 不存在用户名
			return p, proto.Refused_b_u_n_o_p, nil
		}
		p.UserNameLength, p.UserName, _ = s.by2LenNameBE(daByp2)

		if p.UserNameLength < 1 || p.UserName == "" {
			// 不存在用户名
			return p, proto.Refused_b_u_n_o_p, nil
		}

		// 用户名获取完成 必须修改剩余字节
		daByp2 = daByp2[2+p.UserNameLength:]

	}

	// Password Flag (1) 密码标志
	if bs[1] == 49 {
		if len(daByp2) < 3 {
			// 不存在密码
			return p, proto.Refused_b_u_n_o_p, nil
		}
		p.PasswordLength, p.Password, _ = s.by2LenNameBE(daByp2)

		if p.PasswordLength < 1 || p.Password == "" {
			// 不存在密码
			return p, proto.Refused_b_u_n_o_p, nil
		}

		// 最后获取，不用修改字节切片
	}

	return p, proto.Connection_Accepted, nil
}

/*
大端 字节转 长度和名称
失败 返回 false
*/
func (s *MqttDataPack) by2LenNameBE(by []byte) (blen uint16, name string, ok bool) {
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
		p, err = s.unPackPUBLISHProtocol(f)
		return p, err
	}

	switch flag {
	case proto.PUBACK:
		// 解析 PUBACK
		p, err = s.unPackPUBACKProtocol(f)
	case proto.PUBREC:
		// 解析 PUBREC
		p, err = s.unPackPUBRECProtocol(f)
	case proto.PUBREL:
		// 解析 PUBREL
		p, err = s.unPackPUBRELProtocol(f)

	case proto.PUBCOMP:
		//解析 PUBCOMP
		p, err = s.unPackPUBCOMProtocol(f)

	case proto.SUBSCRIBE:

		// 解析订阅协议
		p, err = s.unPackSUBSCRIBEProtocol(f)

	case proto.UNSUBSCRIBE:

		// 解析 取消订阅协议
		p, err = s.unPackUNSUBSCRIBEProtocol(f)

	case proto.PINGREQ:
		//PINGREQ 心跳请求协议
		p = proto.PINGREQProtocol{&proto.Fixed{
			HeaderFlag: flag,
			MsgLen:     0,
		}}
	case proto.DISCONNECT:
		// 断开链接 协议
		p = proto.DISCONNECTProtocol{&proto.Fixed{
			HeaderFlag: flag,
			MsgLen:     0,
		}}

	default:
		return p, errors.New("没有匹配到协议")
	}

	return p, err
}

/*
解包 订阅主题协议
1,创造协议
2，根据 固定报头 解包
*/
func (s *MqttDataPack) unPackSUBSCRIBEProtocol(f *proto.Fixed) (*proto.SUBSCRIBEProtocol, error) {

	// 创造协议
	p := &proto.SUBSCRIBEProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]}, //前两个字节
		TopicFilterList:  []*proto.TopicFilter{},
	}

	// 2 拆解有效载荷
	daBy := f.Data[2:]

	// 一个主题 至少 4个字节 长度2，字符1，Qos1
	for len(daBy) >= 4 {
		tf := &proto.TopicFilter{
			Identifier: 0,
			FilterName: "",
			QoS:        0,
		}

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(daBy[:2]),
			binary.BigEndian, &tf.Identifier)

		tf.FilterName = string(daBy[2:(2 + tf.Identifier)]) // 从 索引[2] 取对应长度值
		tf.QoS = daBy[2+tf.Identifier]                      // Qos

		p.TopicFilterList = append(p.TopicFilterList, tf)

		// 字节截取
		daBy = daBy[2+tf.Identifier+1:]
	}

	// 计算标识符id  大端编码 字节写入数字
	binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
		binary.BigEndian, &p.MsgId)

	return p, nil
}

func (s *MqttDataPack) unPackUNSUBSCRIBEProtocol(f *proto.Fixed) (*proto.UNSUBSCRIBEProtocol, error) {
	// 创造协议
	p := &proto.UNSUBSCRIBEProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]}, //前两个字节
		TopicFilterList:  []*proto.TopicFilter{},
	}
	// 2 拆解有效载荷
	daBy := f.Data[2:]

	// 一个主题 至少 3个字节 长度2，字符1
	for len(daBy) >= 3 {
		tf := &proto.TopicFilter{
			Identifier: 0,
			FilterName: "",
			QoS:        0,
		}

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(daBy[:2]),
			binary.BigEndian, &tf.Identifier)

		tf.FilterName = string(daBy[2:(2 + tf.Identifier)]) // 从 索引[2] 取对应长度值

		p.TopicFilterList = append(p.TopicFilterList, tf)

		// 字节截取
		daBy = daBy[2+tf.Identifier:]
	}

	// 计算标识符id  大端编码 字节写入数字
	binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
		binary.BigEndian, &p.MsgId)

	return p, nil
}

/*
解析发布协议
*/
func (s *MqttDataPack) unPackPUBLISHProtocol(f *proto.Fixed) (*proto.PUBLISHProtocol, error) {

	// 按照 48
	// 创造协议
	p := &proto.PUBLISHProtocol{
		Fixed:            f,
		TopicNameLength:  0, //取可变报头的第二个字节  索引[1]
		TopicName:        "",
		PacketIdentifier: [2]byte{},
		Payload:          nil,

		Qos:    proto.QoS0, // 默认0
		Retain: false,      // 默认0
	}

	binary.Read(bytes.NewBuffer(f.Data[:2]),
		binary.BigEndian, &p.TopicNameLength)

	p.TopicName = string(f.Data[2:(2 + p.TopicNameLength)])

	switch p.Fixed.HeaderFlag {
	case proto.PUBLISH:
		// 48 没有 PacketIdentifier
		p.Payload = f.Data[(2 + p.TopicNameLength):]
		// Qos = 0
		p.Qos = proto.QoS0

	case proto.PUBLISH31:
		// 49 没有 PacketIdentifier 有保留位
		p.Payload = f.Data[(2 + p.TopicNameLength):]
		// Qos = 0
		p.Qos = proto.QoS0
		p.Retain = true

	case proto.PUBLISH32:
		// 50
		p.PacketIdentifier = [2]byte{f.Data[2+p.TopicNameLength], f.Data[2+p.TopicNameLength+1]}
		p.Payload = f.Data[(2 + p.TopicNameLength + 1 + 1):]
		p.Qos = proto.QoS1

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
			binary.BigEndian, &p.MsgId)

	case proto.PUBLISH33:
		// 51
		p.PacketIdentifier = [2]byte{f.Data[2+p.TopicNameLength], f.Data[2+p.TopicNameLength+1]}
		p.Payload = f.Data[(2 + p.TopicNameLength + 1 + 1):]
		p.Qos = proto.QoS1
		p.Retain = true

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
			binary.BigEndian, &p.MsgId)

	case proto.PUBLISH34:
		// 52
		p.PacketIdentifier = [2]byte{f.Data[2+p.TopicNameLength], f.Data[2+p.TopicNameLength+1]}
		p.Payload = f.Data[(2 + p.TopicNameLength + 1 + 1):]
		p.Qos = proto.QoS2

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
			binary.BigEndian, &p.MsgId)

	default:
		// 其他 Qos1,2 都要有标识符
		p.PacketIdentifier = [2]byte{f.Data[2+p.TopicNameLength], f.Data[2+p.TopicNameLength+1]}
		p.Payload = f.Data[(2 + p.TopicNameLength + 1 + 1):]

		// todo 暂定Qos
		p.Qos = proto.QoS2

		// 计算标识符id  大端编码 字节写入数字
		binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
			binary.BigEndian, &p.MsgId)

	}

	return p, nil
}

func (s *MqttDataPack) packPUBREC(p proto.ImplMqttProto) ([]byte, error) {
	p1, ok := p.(*proto.PUBRECProtocol)

	if !ok {
		return nil, errors.New("协议转化字节错误")
	}

	// 固定报头
	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	// 可变报头
	by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])

	return by, nil
}

func (s *MqttDataPack) unPackPUBRELProtocol(f *proto.Fixed) (proto.ImplMqttProto, error) {

	p := &proto.PUBRELProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]},
		MsgId:            0,
	}
	err := binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
		binary.BigEndian, &p.MsgId)

	return p, err

}

func (s *MqttDataPack) packPUBCOMP(p proto.ImplMqttProto) ([]byte, error) {
	p1, ok := p.(*proto.PUBCOMPProtocol)

	if !ok {
		return nil, errors.New("协议转化字节错误")
	}

	// 固定报头
	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	// 可变报头
	by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])

	return by, nil
}

func (s *MqttDataPack) int16ToByBig(ua uint16) []byte {
	var be = make([]byte, 2) // 大端
	// 大端写入  前面的16进制在前 ===》 大端写出
	binary.BigEndian.PutUint16(be, ua)
	return be
}

func (s *MqttDataPack) unPackPUBACKProtocol(f *proto.Fixed) (proto.ImplMqttProto, error) {
	p := &proto.PUBACKProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]},
		MsgId:            0,
	}
	err := binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
		binary.BigEndian, &p.MsgId)

	return p, err
}

func (s *MqttDataPack) unPackPUBRECProtocol(f *proto.Fixed) (proto.ImplMqttProto, error) {
	p := &proto.PUBRECProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]},
		MsgId:            0,
	}
	err := binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
		binary.BigEndian, &p.MsgId)

	return p, err
}

func (s *MqttDataPack) unPackPUBCOMProtocol(f *proto.Fixed) (proto.ImplMqttProto, error) {
	p := &proto.PUBCOMPProtocol{
		Fixed:            f,
		PacketIdentifier: [2]byte{f.Data[0], f.Data[1]},
		MsgId:            0,
	}
	err := binary.Read(bytes.NewBuffer(p.PacketIdentifier[:]),
		binary.BigEndian, &p.MsgId)

	return p, err
}

func (s *MqttDataPack) packPUBREL(p proto.ImplMqttProto) ([]byte, error) {
	p1, ok := p.(*proto.PUBRELProtocol)

	if !ok {
		return nil, errors.New("协议转化字节错误")
	}

	// 固定报头
	by := []byte{p1.GetHeaderFlag()}

	by = append(by, s.msgLenCode(p1.GetMsgLen())...)

	// 可变报头
	by = append(by, p1.PacketIdentifier[0], p1.PacketIdentifier[1])

	fmt.Println("PackPUBREL", p)

	return by, nil
}
