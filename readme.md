轻量级mqtt服务框架
===
# 技术栈
- 日志 使用zap
# 自定义规则
- 遗嘱消息，客户端正常和非正常关闭，订阅者都会收到遗嘱
- 遗嘱信息保留标致，暂时不处理
- 通配符，禁止 $ 的数据发布和订阅， 禁止 "/" 和 ”#“ 发布和订阅

# 链接测试
- mqtt.bijiaox.com
- 端口 1883

# 使用例子
```go
package main

import (
	"github.com/guihai/ghmqtt/mqtt311/server"
)

func main() {
	
	GHmqtt := server.NewGHapi()

	GHmqtt.Run()

}
```