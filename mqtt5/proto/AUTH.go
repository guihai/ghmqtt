package proto

/*
AUTH      = 0xF0    // == 240  1111 0000     C<=>S
*/
type AUTHProtocol struct {
	// 固定报头
	// HeaderFlag uint8    =  AUTH      = 0xF0    // == 240  1111 0000     C<=>S
	// MsgLen
	*Fixed

	// 可变报头
	AuthenticationReasonCode uint8  //
	PropertiesLength         uint32 // 属性长度
	// 属性内容
	AuthenticationMethod string            //认证方法	UTF-8编码字符串
	AuthenticationData   string            // 认证数据	二进制数据
	ReasonString         string            //原因字符串	UTF-8编码字符串
	UserProperty         map[string]string // 用户属性	字符串
}

func NewAUTHProtocolF(f *Fixed) *AUTHProtocol {

	return &AUTHProtocol{
		Fixed:                    f,
		AuthenticationReasonCode: 0,
		PropertiesLength:         0,
		AuthenticationMethod:     "",
		AuthenticationData:       "",
		ReasonString:             "",
		UserProperty:             nil,
	}

}
func NewAUTHProtocol() *AUTHProtocol {

	return &AUTHProtocol{
		Fixed: &Fixed{
			HeaderFlag: AUTH,
			MsgLen:     2, // 默认2
			Data:       nil,
		},
		AuthenticationReasonCode: 0,
		PropertiesLength:         0,
		AuthenticationMethod:     "",
		AuthenticationData:       "",
		ReasonString:             "",
		UserProperty:             nil,
	}

}

func (s *AUTHProtocol) UnPack() error {
	// 可以为0
	if s.Fixed.MsgLen < 1 {
		return nil
	}

	// 读取原因码
	s.AuthenticationReasonCode = s.Fixed.Data[0]

	// 属性长度
	if s.Fixed.MsgLen > 1 {
		// 获取属性
		by := s.Fixed.Data[1:]
		s.PropertiesLength, by = s.unPackPropertyLength(by)

		if s.PropertiesLength > 0 {
			// todo 获取属性
		}
	}

	return nil

}

func (s *AUTHProtocol) Pack() ([]byte, error) {

	// 至少6个字节
	by := make([]byte, 1, 4)
	// 固定报头
	by[0] = s.GetHeaderFlag()

	by = append(by, s.msgLenCode(s.GetMsgLen())...)

	// 可变报头
	//原因码
	by = append(by, s.AuthenticationReasonCode)

	// todo 属性默认0
	by = append(by, s.msgLenCode(s.PropertiesLength)...)

	return by, nil

}
