package proto

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

func NewPINGREQProtocol(f *Fixed) *PINGREQProtocol {
	return &PINGREQProtocol{f}
}
