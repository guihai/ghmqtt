package proto

const (
	CONNECT = 0x10 //  == 16   0001 0000        C=>S
	CONNACK = 0x20 // == 32    0010 0000        S=>C
	PUBLISH = 0x30 // == 48   0011 xxxx   大于等于48 小于 64   C<=>S

	PUBLISH31 = 0x31 // == 49   0011 0001   保留信息    C<=>S

	PUBLISH32 = 0x32 // == 50   0011 0010   需要 PUBACK   C<=>S

	PUBLISH33 = 0x33 // == 51   0011 0011   需要 PUBACK 保留信息   C<=>S

	PUBLISH34 = 0x34 // == 52   0011 0100   Qos2 需要 PUBREC    C<=>S

	PUBLISHMAX = 0x3D // == 61   订阅协议二进制转化 ，最大值

	PUBACK      = 0x40 // == 64   0100 0000         C<=>S
	PUBREC      = 0x50 // == 80   0101 0000         C<=>S
	PUBREL      = 0x62 //  ==98   0110 0010         C<=>S
	PUBCOMP     = 0x70 //  ==112       0111 0000       C<=>S
	SUBSCRIBE   = 0x82 //  == 130        1000 0010     C=>S
	SUBACK      = 0x90 // ==144            1001 0000    S=>C
	UNSUBSCRIBE = 0xA2 // == 162        1010 0010      C=>S
	UNSUBACK    = 0xB0 // == 176    1011 0000          S=>C
	PINGREQ     = 0xC0 // ==  192     1100 0000       C=>S
	PINGRESP    = 0xD0 //  == 208      1101 0000      S=>C
	DISCONNECT  = 0xE0 // == 224    1110 0000         C=>S
)

const (
	//CONNACK_RETURNCODE

	Connection_Accepted uint8 = iota // Connection Accepted 链接已接受
	Refused_u_p_v                    // unacceptable protocol version
	Refused_i_r                      // Refused, identifier rejected
	Refused_S_u                      // Refused, Server unavailable
	Refused_b_u_n_o_p                // Refused, bad user name or password
	Refused_n_a                      //  Refused, not authorized
)

const (
	//QoS 服务质量要求

	QoS0 uint8 = iota //QoS = 0 – 最多发一次
	QoS1              //QoS = 1 – 最少发一次
	QoS2              //QoS = 2 – 保证收一次

	Failure = 0x80 // == 128 订阅确认失败
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

// 可变报头，部分控制报文包含 通用型
type Variable struct {
	ProtoNameLen uint16 // 两个字节
	ProtoName    string // 根据ProtoNameLen 长度获取的数据
	Version      uint8  // 一个字节 协议级别 版本号 3.11 = 0x04
	ConnectFlag  uint8  // 一个字节  链接标志 可以控制有效载荷
	KeepAlive    uint8  // 两个字节 保持连接 Keep Alive MSB 和 保持连接 Keep Alive LSB
}

/*
链接协议结构体  客户端到服务端
CONNECT = 0x10
*/
type CONNECTProtocol struct {
	// 固定报头
	*Fixed
	// 可变报头
	ProtoNameLen uint16 // 两个字节  大端解码返回数值
	ProtoName    string // 根据ProtoNameLen 长度获取的数据
	Version      uint8  // 一个字节 协议级别 版本号
	ConnectFlag  uint8  // 一个字节  链接标志 可以控制有效载荷
	KeepAlive    uint16 // 两个字节 大端解码返回数值

	// 有效载荷 根据可变报头参数  ConnectFlag 这里的数据有变化
	//客户端标识符 (ClientId) 必须存在而且必须是 CONNECT 报文有效载荷的第一个字段
	ClientIDLength uint16 // 两个字节  大端解码返回数值  可以设定设定长度不要超过 200
	ClientID       string // 根据长度 获取数据

	//如果可变报头连接标志部分遗嘱标志被设置为 1，则有效载荷的下一个字段是遗嘱主题（Will Topic）
	WillTopicLength uint16 // 大端解码返回数值
	WillTopic       string

	// 如果可变报头连接标志部分遗嘱标志被设置为 1，有效载荷的下一个字段是遗嘱消息。
	WillMessageLength uint16
	WillMessage       string

	// 如果可变报头连接标志部分用户名（User Name）标志被设置为 1，有效载荷的下一个字段就是它。
	UserNameLength uint16
	UserName       string

	// 如果可变报头连接标志部分密码（Password）标志被设置为 1，有效载荷的下一个字段就是它。
	PasswordLength uint16
	Password       string

	// 以下数据解析出使用
	// 解析出 cleansession
	CleanSession bool // 默认 true 业务中处理  服务器必须在客户端断开之后继续存储/保持客户端的订阅状态
	// 遗嘱标识
	WillFlag   bool
	WillRetain bool
	WillQos    uint8
}

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

}

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

	// 有效载荷  剩余字节都是主题内容
	Payload []byte

	// 解析Qos 用于处理
	Qos uint8

	// 解析 Retain 用于处理
	Retain bool // 0 或者1

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储
}

/*
PUBACK –发布确认 协议
PUBACK 报文是对 QoS 1 等级的 PUBLISH 报文的响应。
发送的 PUBLISH 报文必须包含报文标识符且 QoS 等于 1，DUP 等于 0
PUBACK      = 0x40 // == 64   0100 0000         C<=>S
*/
type PUBACKProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PUBACK      = 0x40 // == 64   0100 0000         C<=>S
	// MsgLen = 2
	*Fixed

	// 可变报头
	// 等于需要确认的 PUBLISH 的  PacketIdentifier
	PacketIdentifier [2]byte // 两个字节  第一个标识 MSB   和 第二个标识LSB

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

}

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

}

/*
发布释放 协议  Qos2 第二步
PUBREL      = 0x62 //  ==98   0110 0010         C<=>S
*/
type PUBRELProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PUBREL      = 0x62 //  ==98   0110 0010         C<=>S
	// MsgLen = 2
	*Fixed

	// 可变报头
	// 等于需要确认的 PUBLISH 的  PacketIdentifier
	PacketIdentifier [2]byte // 两个字节

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

}

/*
发布完成协议  Qos2 最后响应
PUBCOMP     = 0x70 //  ==112       0111 0000       C<=>S
*/
type PUBCOMPProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PUBCOMP     = 0x70 //  ==112       0111 0000       C<=>S
	// MsgLen = 2
	*Fixed

	// 可变报头
	// 等于需要确认的 PUBLISH 的  PacketIdentifier
	PacketIdentifier [2]byte // 两个字节

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

}

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

	// 有效载荷
	TopicFilterList []*TopicFilter

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储
}

// SUBSCRIBEProtocol 使用
type TopicFilter struct {
	Identifier uint16 // 两个字节
	FilterName string // 根据Identifier 长度获取
	QoS        uint8  // 本组最后一个字节 只能 从常量  QoS0,1,2中取值
}

/*
订阅确认 协议
SUBACK      = 0x90 // ==144            1001 0000    S=>C
PacketIdentifier 必须和SUBSCRIBE 相同
*/
type SUBACKProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  SUBACK      = 0x90 // ==144            1001 0000    S=>C
	//	MsgLen = 1-4
	*Fixed

	// 可变报头
	PacketIdentifier [2]byte // 两个字节 和订阅相同

	// 有效载荷   使用常量Qos0,1,2  或者 Failure
	ReturnCodeList []uint8

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储
}

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

	// 有效载荷   没有 QoS 所以至少是3个字节
	TopicFilterList []*TopicFilter

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储
}

/*
取消订阅确认 协议
UNSUBACK    = 0xB0 // == 176    1011 0000          S=>C
*/
type UNSUBACKProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  UNSUBACK    = 0xB0 // == 176    1011 0000          S=>C
	//	MsgLen = 2
	*Fixed
	// 可变报头
	PacketIdentifier [2]byte // 两个字节  和取消订阅相同

	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

}

/*
PINGREQ 心跳请求协议
PINGREQ     = 0xC0 // ==  192     1100 0000       C=>S

*/
type PINGREQProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PINGREQ     = 0xC0 // ==  192     1100 0000       C=>S
	//	MsgLen = 0  必须为 0
	*Fixed
}

/*
PINGRESP  心跳响应协议
PINGRESP    = 0xD0 //  == 208      1101 0000      S=>C

*/
type PINGRESPProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  PINGRESP    = 0xD0 //  == 208      1101 0000      S=>C
	//	MsgLen = 0  必须为 0
	*Fixed
}

/*
客户端断开链接协议
DISCONNECT  = 0xE0 // == 224    1110 0000         C=>S
*/
type DISCONNECTProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  DISCONNECT  = 0xE0 // == 224    1110 0000         C=>S
	//	MsgLen = 0  必须为 0
	*Fixed
}

// 遗嘱消息结构体
type Will struct {
	WillTopic   string
	WillMessage string
	WillRetain  bool
	WillQos     uint8
}
