mqtt5
===
# 使用案例
```go
package main

import (
	"github.com/guihai/ghmqtt/mqtt5/server"
)

func main() {

	GHmqtt := server.NewGHapi()

	GHmqtt.Run()

}
```