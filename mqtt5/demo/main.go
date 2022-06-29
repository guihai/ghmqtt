package main

import (
	"hshiye/tvl2mqtt/mqtt5/demo/router"
	"hshiye/tvl2mqtt/mqtt5/proto"
	"hshiye/tvl2mqtt/mqtt5/server"
)

func main() {

	GHmqtt := server.NewGHapi()

	// 注册链接验证
	GHmqtt.SetConnectVerify(router.CheckConn)

	// 注册路由
	//addRouter(GHmqtt)

	// 开启 http 服务

	GHmqtt.Run()

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
