package server

import (
	"github.com/guihai/ghmqtt/mqtt5/proto"
	"github.com/guihai/ghmqtt/utils"
	"github.com/guihai/ghmqtt/utils/zaplog"
	"go.uber.org/zap"
	"hash/fnv"
)

/*
每种协议，都有对应的路由处理
每种协议，可以设置 前置中间件，路由业务，后置中间件

根据发布和订阅主题不同，这个业务在业务层处理，不在基础服务层

*/
type RouterManager struct {

	//链接路由 map[协议类型编号]ImplBaseRouter  参考 proto.CONNECT 等常量
	routerMap map[uint8]ImplBaseRouter

	// 链接验证方法，用户名等信息认证在这里
	connectVerify ConnectVerifyFUNC

	// 创建协程池 工作单位+任务队列
	// 工作单位数量
	workPoolSize uint32
	// 任务队列 管道 切面，数据就是请求
	taskQueue []chan *Request

	// WorkPoolIsOn
	poolOn bool
}

func newRouterManager() *RouterManager {

	r := &RouterManager{

		routerMap: make(map[uint8]ImplBaseRouter),

		// 实现一个默认的ConnectVerifyFUNC
		connectVerify: func(protocol *proto.CONNECTProtocol) uint8 {
			return 0
		},

		// 创建协程池
		workPoolSize: utils.GO.WorkPoolSize,
		//一个worker对应一个queue
		taskQueue: make([]chan *Request, utils.GO.WorkPoolSize),

		poolOn: false, // 协程池未启动
	}

	// 实现默认的链接路由器
	r.addRouter(proto.CONNECT, &CONNECTRouter{})
	// 实现默认的断开
	r.addRouter(proto.DISCONNECT, &DISCONNECTRouter{})
	// 默认的 心跳路由
	r.addRouter(proto.PINGREQ, &PINGREQRouter{})
	// 默认的订阅 路由
	r.addRouter(proto.SUBSCRIBE, &SUBSCRIBERouter{})
	// 默认取消订阅 路由
	r.addRouter(proto.UNSUBSCRIBE, &UNSUBSCRIBERouter{})

	//  默认发布协议  48 路由
	r.addRouter(proto.PUBLISH, &PUBLISHRouter{})
	// 发布协议 49 路由 有保留信息标志
	r.addRouter(proto.PUBLISH31, &PUBLISHRouter{})

	// 发布协议 50 路由
	r.addRouter(proto.PUBLISH32, &PUBLISHRouter{})
	// 发布协议 51 路由
	r.addRouter(proto.PUBLISH33, &PUBLISHRouter{})

	// 发布协议 52 路由
	r.addRouter(proto.PUBLISH34, &PUBLISHRouter{})

	// 发布消息响应路由
	r.addRouter(proto.PUBACK, &PUBACKRouter{})

	// 释放消息协议 Qos2 流程
	r.addRouter(proto.PUBREL, &PUBRELRouter{})

	//
	r.addRouter(proto.PUBREC, &PUBRECRouter{})

	r.addRouter(proto.PUBCOMP, &PUBCOMPRouter{})

	return r

}

/*
设置链接验证器
*/
func (s *RouterManager) setConnectVerify(cvf ConnectVerifyFUNC) {
	s.connectVerify = cvf
}

/*
添加路由
默认会添加断开路由，所以断开理由可以重新添加，后面的会覆盖
添加路由先验证是否存在
*/
func (s *RouterManager) addRouter(i uint8, router ImplBaseRouter) {

	// 新路由会覆盖默认路由
	s.routerMap[i] = router
	//fmt.Println("路由添加成功 = ", i)
}

/*
根据请求获取路由执行方法
1 检查是否存在
2 执行前置，业务，后置方法
*/

func (s *RouterManager) doRouterFunc(request *Request) {

	r, ok := s.routerMap[request.proto.GetHeaderFlag()]
	if !ok {
		zaplog.ZapLogger.Warn("没有路由 协议 = ", zap.Uint8("协议编号", request.proto.GetHeaderFlag()))
		return
	}

	// 前置方法
	r.PreHandle(request)
	// 业务方法
	r.Handle(request)
	// 后置方法
	r.PostHandle(request)

}

/*
协程池是否开启
*/
func (s *RouterManager) workPoolIsOn() bool {
	return s.poolOn
}

/*
初始化协程池，服务已开启就要启动
*/
func (s *RouterManager) startWorkerPool() {

	for i := 0; i < int(s.workPoolSize); i++ {

		// 初始化任务队列的管道
		s.taskQueue[i] = make(chan *Request, utils.GO.TaskQueueMaxSize)

		// 启动一个工作单位
		go s.startWorker(i, s.taskQueue[i])
	}

	s.poolOn = true
}

/*
开启工作单位
*/
func (s *RouterManager) startWorker(i int, requests chan *Request) {
	zaplog.ZapLogger.Info("【路由管理者协程池开启】 创建工作者", zap.Int("编号", i))
	//不断的等待队列中的消息,然后进行路由处理
	for {
		select {
		case re := <-requests:
			s.doRouterFunc(re)
		}
	}

}

//将消息交给TaskQueue,由worker进行处理
func (s *RouterManager) sendReqToTaskQueue(request *Request) {
	// 采用 链接id 取余数 放入对应的消息队列
	cid := request.getConn().getClientID()
	//pid := request.Proto.GetHeaderFlag()

	wid := dHash(cid) % s.workPoolSize

	//fmt.Println("添加 client =", cid, " 请求 pid =", pid, "到 workerID=", wid)

	s.taskQueue[wid] <- request
}

// 简易hash 算法
func dHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))

	return h.Sum32()
}
