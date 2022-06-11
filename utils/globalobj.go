package utils

import (
	"encoding/json"
	"github.com/guihai/ghmqtt/utils/zaplog"
	"io/ioutil"
)

/*
全局配置对象
*/

type GlobalObj struct {
	// 服务名称
	Name string
	// 绑定ip
	IP string
	// 绑定端口
	Port uint16
	// 传输协议
	Tcp string
	// 最大连接数
	MaxConn uint32
	// 链接活跃时长，秒，超过时长不活跃会关闭链接
	ConnLiveTime uint8
	// 数据包最大值
	MaxPacketSize uint32
	// 版本号
	Version string

	// 协程池
	WorkPoolSize uint32
	// 协程池任务队列的最大容量
	TaskQueueMaxSize uint32

	// 日志配置
	LogCfg *zaplog.LogConfig

	// 客户端参数
	MQTTClient *MQTTClient
}

/*
客户端参数设置
*/
type MQTTClient struct {
	ClientIDLen uint16 // 设定长度，默认最长65535
	NameLen     uint16 // 设定长度，默认最长65535
	PSDLen      uint16 // 设定长度，默认65535

	AdminName string // utf8 编码
	AdminPSD  string // 密码

}

// 定义全局使用的变量
var GO *GlobalObj

// 初始化 全局变量，使用 init 方法 ，包被调用的时候会先调用这个方法
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GO = &GlobalObj{
		Name:          "GHMQTT",
		IP:            "0.0.0.0",
		Port:          1883,
		Tcp:           "tcp4",
		MaxConn:       100,
		MaxPacketSize: 2048,
		Version:       "v1.0",
		ConnLiveTime:  120, // 默认120秒

		// 协程池
		WorkPoolSize: 10, // 和cpu 数量匹配合适
		// 协程池任务队列的最大容量
		TaskQueueMaxSize: 1024,

		LogCfg: &zaplog.LogConfig{
			Filename:   "./log/logs.json",
			MaxSize:    128,
			MaxAge:     7,
			MaxBackups: 7,
			Compress:   true,
			Level:      0,
			StdoutFLag: true,
		},

		MQTTClient: &MQTTClient{
			ClientIDLen: 200,
			NameLen:     200,
			PSDLen:      200,
			AdminName:   "admin",
			AdminPSD:    "0608",
		},
	}

	// 加载配置参数
	//reloadConfig()

	zaplog.InitLogger(GO.LogCfg)

}

func reloadConfig() {
	buf, err := ioutil.ReadFile("etc/etc.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(buf, GO)
	if err != nil {
		panic(err)
	}
}
