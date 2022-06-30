package server

import (
	"fmt"
	"github.com/guihai/ghmqtt/utils"
	"github.com/guihai/ghmqtt/utils/zaplog"
	"go.uber.org/zap"
	"net"
)

type Server struct {
	name string // 服务名称
	ip   string // 监听ip 默认 0.0.0.0
	port uint16 // 监听端口 默认  18883
	tcp  string // 传输协议 默认 tcp4

	// 结束服务信号
	exitChan chan bool

	// 链接对象管理器
	connMer *ConnManager

	// 路由管理器
	routerMer *RouterManager

	// 主题管理器 服务启动就要启动主题列表
	topicMer *TopicManager
}

func newServer() *Server {
	ser := &Server{
		name:      utils.GO.Name,
		ip:        utils.GO.IP,
		port:      utils.GO.Port,
		tcp:       utils.GO.Tcp,
		connMer:   newConnManager(),
		routerMer: newRouterManager(),

		exitChan: make(chan bool),
	}
	// 开启 自带的主题管理器标识
	ser.topicMer = newTopicManager(ser)

	return ser
}

// 实现 接口方法
/*
启动tcp 服务
*/
func (s *Server) start() {
	zaplog.ZapLogger.Info("【启动服务】" + s.name)

	// 开启协程池，等待工作
	s.routerMer.startWorkerPool()

	// 开启协程监听
	go func() {
		// 0 创建地址
		addr, err := net.ResolveTCPAddr(s.tcp, fmt.Sprintf("%s:%d", s.ip, s.port))
		if err != nil {
			zaplog.ZapLogger.Warn("【失败】获取tcp服务地址失败：" + err.Error())
			return
		}
		// 1，启动监听
		lis, err := net.ListenTCP(s.tcp, addr)
		if err != nil {
			zaplog.ZapLogger.Warn("【失败】启动监听失败，" + err.Error())

			panic(err)
		}

		zaplog.ZapLogger.Info("【服务开启成功】", zap.String("name", s.name), zap.Uint16("port", s.port))

		// 2 开启循环接收链接
		for {
			conn, err := lis.AcceptTCP()
			if err != nil {
				zaplog.ZapLogger.Error("【错误】，接收连接错误" + err.Error())
				// 连接失败，继续下一个链接
				continue
			}

			// 封装链接对象
			co := newConn(conn, s)

			go co.start()
		}
	}()

}

// 关闭服务
// todo 关闭所有资源检查
func (s *Server) stop() {

	zaplog.ZapLogger.Info("【服务开始关闭】")

	// 清理所有链接
	s.connMer.clearConn()

	// topic 关闭所有资源
	s.topicMer.stop()

	zaplog.ZapLogger.Info("【服务关闭】" + s.name + "停止服务，再见")

	// 关闭通道赋值
	s.exitChan <- true

}

// 实现 接口方法
func (s *Server) run() {

	//调用start
	s.start()
	// 开启阻塞，
	select {
	case <-s.exitChan:
		// 结束阻塞，关闭服务
		return
	}
}

/*
注册路由
*/

//func (s *Server) addRouter(i uint8, router ImplBaseRouter) {
//	s.RouterMer.AddRouter(i, router)
//}

///*
//注册链接验证
//*/
//func (s *Server) setConnectVerify(cvf ConnectVerifyFUNC) {
//	s.RouterMer.SetConnectVerify(cvf)
//}
