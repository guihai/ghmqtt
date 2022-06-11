package proto

type ImplMqttProto interface {
	GetHeaderFlag() uint8
	GetMsgLen() uint32
	GetData() []byte
}
