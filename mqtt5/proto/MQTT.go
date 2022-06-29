package proto

// 协议编号
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
	AUTH        = 0xF0 // == 240  1111 0000     C<=>S
)

// 响应码
const (
	//CONNACK_RETURNCODE

	Success            uint8 = 0    // Connection Accepted 链接已接受
	Unspecified_error  uint8 = 0x80 // 未指明的错误	服务端不愿透露的错误，或者没有适用的原因码。
	Malformed_Packet   uint8 = 0x81 // 无效报文	CONNECT报文内容不能被正确的解析。
	Protocol_Error     uint8 = 0x82 // 协议错误	CONNECT报文内容不符合本规范。
	Implementation_s_e uint8 = 0x83 // 实现特定错误	CONNECT有效，但不被服务端所接受。
	UnsupportedPV      uint8 = 0x84 // 协议版本不支持	服务端不支持客户端所请求的MQTT协议版本
	ClientInotv        uint8 = 0x85 // 客户标识符无效	客户标识符有效，但未被服务端所接受。
	BadUNorP           uint8 = 0x86 //用户名密码错误	客户端指定的用户名密码未被服务端所接受。
	Notauthorized      uint8 = 0x87 // 未授权	客户端未被授权连接。
	Server_unavailable uint8 = 0x88 // 服务端不可用	MQTT服务端不可用。
	Server_busy        uint8 = 0x89 // 服务端正忙	服务端正忙，请重试。
	Banned             uint8 = 0x8A // 客户端被禁止，请联系服务端管理员。
	Bad_a_method       uint8 = 0x8C // 无效的认证方法	认证方法未被支持，或者不匹配当前使用的认证方法。
	Topic_Name_invalid uint8 = 0x90 // 主题名无效	遗嘱主题格式正确，但未被服务端所接受。
	Packet_too_large   uint8 = 0x95 //报文过长	CONNECT报文超过最大允许长度。
	Quota_exceeded     uint8 = 0x97 // 超出配额	已超出实现限制或管理限制。
	Payload_f_i        uint8 = 0x99 // 载荷格式无效	遗嘱载荷数据与载荷格式指示符不匹配。
	Retain_n_s         uint8 = 0x9A //不支持保留	遗嘱保留标志被设置为1，但服务端不支持保留消息。
	QoS_n_s            uint8 = 0x9B //不支持的QoS等级	服务端不支持遗嘱中设置的QoS等级。
	Use_a_s            uint8 = 0x9C //（临时）使用其他服务端	客户端应该临时使用其他服务端。
	Server_moved       uint8 = 0x9D //服务端已（永久）移动	客户端应该永久使用其他服务端
	Connection_r_e     uint8 = 0x9F // 超出连接速率限制	超出了所能接受的连接速率限制。

	// 以下是非connack 使用的
	Disconnect_w_W_M   uint8 = 0x04 // 包含遗嘱的断开	DISCONNECT
	No_m_subscribers   uint8 = 0x10 // 无匹配订阅	PUBACK, PUBREC
	No_subscription_e  uint8 = 0x11 //订阅不存在	UNSUBACK
	Continue_a         uint8 = 0x18 //继续认证	AUTH
	Re_authenticate    uint8 = 0x19 // 重新认证	AUTH
	Server_s_down      uint8 = 0x8B //服务端关闭中	DISCONNECT
	Keep_Alive_to      uint8 = 0x8D // 保活超时	DISCONNECT
	Session_to         uint8 = 0x8E //会话被接管	DISCONNECT
	Topic_Filter_i     uint8 = 0x8F // 主题过滤器无效	SUBACK, UNSUBACK, DISCONNECT
	Packet_Iinuse      uint8 = 0x91 //报文标识符已被占用	PUBACK, PUBREC, SUBACK, UNSUBACK
	Packet_Inotf       uint8 = 0x92 //报文标识符无效	PUBREL, PUBCOMP
	Receive_M_e        uint8 = 0x93 //接收超出最大数量	DISCONNECT
	Topic_A_i          uint8 = 0x94 // 主题别名无效	DISCONNECT
	Message_r_t_h      uint8 = 0x96 //消息太过频繁	DISCONNECT
	Administrative_a   uint8 = 0x98 //管理行为	DISCONNECT
	Shared_S_n_s       uint8 = 0x9E //不支持共享订阅	SUBACK, DISCONNECT
	Maximum_c_t        uint8 = 0xA0 //最大连接时间
	Subscription_I_n_s uint8 = 0xA1 //不支持订阅标识符	SUBACK, DISCONNECT
	Wildcard_S_n_s     uint8 = 0xA2 //不支持通配符订阅	SUBACK, DISCONNECT
)

const (
	//QoS 服务质量要求

	QoS0 uint8 = iota //QoS = 0 – 最多发一次
	QoS1              //QoS = 1 – 最少发一次
	QoS2              //QoS = 2 – 保证收一次

)

// 内置属性编号
const (
	//UTF-8编码字符串  两个字节标识长度，后面的是内容
	PayloadFI       uint8 = 0x01 //Payload Format Indicator 载荷格式说明	字节	  PUBLISH, Will Properties
	MessageEI       uint8 = 0x02 //Message Expiry Interval 消息过期时间	四字节整数	PUBLISH, Will Properties
	ContentType     uint8 = 0x03 //Content Type 内容类型	UTF-8编码字符串	PUBLISH, Will Properties
	ResponseTopic   uint8 = 0x08 //Response Topic  响应主题	UTF-8编码字符串	PUBLISH, Will Properties
	CorrelationData uint8 = 0x09 //Correlation Data 相关数据	二进制数据	PUBLISH, Will Properties
	SubscriptionI   uint8 = 0x0B //Subscription Identifier 定义标识符	变长字节整数	PUBLISH, SUBSCRIBE
	SessionEI       uint8 = 0x11 //Session Expiry Interval 会话过期间隔	四字节整数	CONNECT, CONNACK, DISCONNECT
	AssignedCI      uint8 = 0x12 //Assigned Client Identifier 分配客户标识符	UTF-8编码字符串	CONNACK
	ServerKA        uint8 = 0x13 //Server Keep Alive 服务端保活时间	双字节整数	CONNACK
	AuthenticationM uint8 = 0x15 // Authentication Method 认证方法	UTF-8编码字符串	CONNECT, CONNACK, AUTH
	AuthenticationD uint8 = 0x16 //Authentication Data 认证数据	二进制数据	CONNECT, CONNACK, AUTH
	RequestPI       uint8 = 0x17 //Request Problem Information 请求问题信息	字节	CONNECT
	WillDI          uint8 = 0x18 // Will Delay Interval 遗嘱延时间隔	四字节整数	Will Properties
	RequestRI       uint8 = 0x19 //Request Response Information 请求响应信息	字节	CONNECT
	ResponseI       uint8 = 0x1A // Response Information 请求信息	UTF-8编码字符串	CONNACK
	ServerRef       uint8 = 0x1C // Server Reference 服务端参考	UTF-8编码字符串	CONNACK, DISCONNECT
	ReasonString    uint8 = 0x1F // Reason String 原因字符串	UTF-8编码字符串	CONNACK, PUBACK, PUBREC, PUBREL, PUBCOMP, SUBACK, UNSUBACK, DISCONNECT, AUTH
	ReceiveMaximum  uint8 = 0x21 // Receive Maximum 接收最大数量	双字节整数	CONNECT, CONNACK
	TopicAM         uint8 = 0x22 // Topic Alias Maximum 主题别名最大长度	双字节整数	CONNECT, CONNACK
	TopicAlias      uint8 = 0x23 // Topic Alias 主题别名	双字节整数	PUBLISH
	MaximumQoS      uint8 = 0x24 // Maximum QoS 最大QoS	字节	CONNACK
	RetainA         uint8 = 0x25 //Retain Available   字节	CONNACK
	UserProperty    uint8 = 0x26 //User Property 用户属性	UTF-8字符串对	CONNECT, CONNACK, PUBLISH, Will Properties, PUBACK, PUBREC, PUBREL, PUBCOMP, SUBSCRIBE, SUBACK, UNSUBSCRIBE, UNSUBACK, DISCONNECT, AUTH
	MaximumPS       uint8 = 0x27 //Maximum Packet Size 最大报文长度	四字节整数	CONNECT, CONNACK
	WildcardSA      uint8 = 0x28 // Wildcard Subscription Available 通配符订阅可用性	字节	CONNACK
	SubscriptionIA  uint8 = 0x29 //Subscription Identifier Available 订阅标识符可用性	字节	CONNACK
	SharedSA        uint8 = 0x2A //Shared Subscription Available 共享订阅可用性	字节	CONNACK
)
