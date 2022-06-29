package proto

type ImplMqttProto interface {
	GetHeaderFlag() uint8
	GetMsgLen() uint32
	GetData() []byte
	GetAckCode() uint8

	by2LenNameBE(by []byte) (blen uint16, name string, ok bool)
	by2Len32(by []byte) uint32
	msgLenEnCode(by []byte) uint32
	msgLenCode(sln uint32) []byte
	int16ToByBig(ua uint16) []byte
	unPackPropertyLength(by []byte) (uint32, []byte)

	// 封包
	Pack() ([]byte, error)
	// 解包 需要返回 响应码，或者错误
	UnPack() error
}
