package main

import (
	"github.com/gin-gonic/gin"
	"github.com/guihai/ghmqtt/mqtt311/demo/router"
	"github.com/guihai/ghmqtt/mqtt311/proto"
	"github.com/guihai/ghmqtt/mqtt311/server"
)

func main() {

	GHmqtt := server.NewGHapi()

	// 注册链接验证
	GHmqtt.SetConnectVerify(router.CheckConn)

	// 注册路由
	addRouter(GHmqtt)

	// 开启 http 服务
	go httpRun()

	router.MQTTAPI = GHmqtt

	GHmqtt.Run()

}

// 开启 http 服务进行管理
func httpRun() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.Recovery())

	router.RouterInit(r)

	r.Run(":31180") // 监听并在 0.0.0.0:8080 上启动服务
}

//注册路由
func addRouter(gh *server.GHapi) {

	gh.AddRouter(proto.CONNECT, &router.CONNECTRouter{})
	gh.AddRouter(proto.DISCONNECT, &router.DISCONNECTRouter{})
	gh.AddRouter(proto.PINGREQ, &router.PINGREQRouter{})
	gh.AddRouter(proto.SUBSCRIBE, &router.SUBSCRIBERouter{})
	gh.AddRouter(proto.UNSUBSCRIBE, &router.UNSUBSCRIBERouter{})
	gh.AddRouter(proto.PUBLISH, &router.PUBLISHRouter{})
	gh.AddRouter(proto.PUBLISH31, &router.PUBLISHRouter{})
	gh.AddRouter(proto.PUBLISH32, &router.PUBLISHRouter{})
	gh.AddRouter(proto.PUBLISH33, &router.PUBLISHRouter{})
	gh.AddRouter(proto.PUBLISH34, &router.PUBLISHRouter{})
	gh.AddRouter(proto.PUBACK, &router.PUBACKRouter{})
	gh.AddRouter(proto.PUBREL, &router.PUBRELRouter{})
	gh.AddRouter(proto.PUBREC, &router.PUBRECRouter{})
	gh.AddRouter(proto.PUBCOMP, &router.PUBCOMPRouter{})
}
