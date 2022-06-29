package types

// 发布消息请求
type PublishMsg struct {

	// 主题
	TopicName string
	// 内容
	TopicMsg string
	// 保留标识
	Retain bool
	// Qos级别
	Qos uint8
}
