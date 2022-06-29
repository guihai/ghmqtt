package proto

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

func NewPINGRESPProtocol() *PINGRESPProtocol {
	return &PINGRESPProtocol{
		Fixed: &Fixed{
			HeaderFlag: PINGRESP,
			MsgLen:     0,
		},
	}
}

func (s *PINGRESPProtocol) Pack() ([]byte, error) {
	by := make([]byte, 2, 2)

	by[0] = s.Fixed.HeaderFlag
	by[1] = 0

	return by, nil
}
