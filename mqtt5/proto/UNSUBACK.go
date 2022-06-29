package proto

/*
取消订阅确认 协议
UNSUBACK    = 0xB0 // == 176    1011 0000          S=>C
*/
type UNSUBACKProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  UNSUBACK    = 0xB0 // == 176    1011 0000          S=>C
	//	MsgLen = 1-4
	*Fixed
	// 可变报头
	PacketIdentifier [2]byte // 两个字节  和取消订阅相同
	// 报文标识符转化为 MsgId
	MsgId uint16 // PacketIdentifier = MsgID 大端存储

	// 属性
	PropertiesLength uint32            // 1-4 字节 和 msgLen 相同
	ReasonString     string            //原因字符串	UTF-8编码字符串
	UserProperty     map[string]string // 用户属性	字符串

	// 有效载荷   使用响应码 从订阅处理后获取
	ReturnCodeList []uint8
}

func NewUNSUBACKProtocol(tfl uint32, pid [2]byte, rcl []uint8) *UNSUBACKProtocol {
	p := &UNSUBACKProtocol{
		Fixed: &Fixed{
			HeaderFlag: UNSUBACK,
			MsgLen:     2 + 1 + tfl, // 标识符 2 字节，属性1 返回code len(sp.TopicFilterList)
			Data:       nil,
		},
		// 和订阅协议相同
		PacketIdentifier: pid,
		// 全部返回订阅成功
		ReturnCodeList:   rcl,
		PropertiesLength: 0, // 默认0
	}
	return p
}

func (s *UNSUBACKProtocol) Pack() ([]byte, error) {

	// 至少6个字节
	by := make([]byte, 1, 6)
	by[0] = s.GetHeaderFlag()

	by = append(by, s.msgLenCode(s.GetMsgLen())...)

	by = append(by, s.PacketIdentifier[0], s.PacketIdentifier[1])

	// 属性长度
	// todo 属性默认0
	by = append(by, s.msgLenCode(s.PropertiesLength)...)

	by = append(by, s.ReturnCodeList...)

	return by, nil
}
