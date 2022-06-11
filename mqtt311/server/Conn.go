package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/guihai/ghmqtt/mqtt311/proto"
	"github.com/guihai/ghmqtt/utils"
	"github.com/guihai/ghmqtt/utils/zaplog"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type Conn struct {
	netConn  *net.TCPConn // 原始测tcp 链接对象
	clientID string       // mqtt 生成 clientid  在第一次链接的时候获取
	isClose  bool         // 是否关闭链接   true 标识关闭，false 未关闭

	writerBuffChan chan []byte // 写数据通道 有缓冲

	// 上下文管理 管理关闭
	ctx context.Context
	cal context.CancelFunc

	//当前Conn属于哪个Server 方便调用server中的 链接管理器和路由管理器
	ofServer *Server

	// 链接属性，方便业务中使用
	keyValue map[string]interface{}
	//保护链接属性修改的锁
	kvLock sync.RWMutex

	// 超时关闭链接设置
	// 活跃在线通道，客户请求后 将true置入此通道  无缓冲阻塞
	liveChan chan bool
	// 活跃时间设置  单位秒  1，可以设置参数；2可以和客户端的live时间 成比例
	liveTime uint8
}

func newConn(conn *net.TCPConn, ser *Server) *Conn {
	c := &Conn{
		netConn: conn,
		isClose: false,

		ofServer: ser,

		// 有缓冲写入通道
		writerBuffChan: make(chan []byte, utils.GO.MaxPacketSize), // 返回写数据通道

		// 初始化属性
		keyValue: make(map[string]interface{}),

		// 初始化活跃通道
		liveChan: make(chan bool),
		liveTime: utils.GO.ConnLiveTime,
	}
	return c
}

func (s *Conn) start() {

	// 开启上下文 管理
	s.ctx, s.cal = context.WithCancel(context.Background())

	// 第一次链接 要设置 client 仅设置一次
	err := s.setClientID()
	if err != nil {
		zaplog.ZapLogger.Error("【错误】连接启动错误" + err.Error())
		// 关闭链接
		s.finalStop()
		return
	}

	// 读数据
	go s.read()

	// 写数据
	go s.write()

	// 开启监听 上下文是否关闭
	for {
		select {
		case <-s.ctx.Done(): // 上下文关闭了
			// 执行最终关闭链接
			s.finalStop()

			return
		case <-s.liveChan:
			// 活跃通道获取数据，不操作 执行下一次循环
			continue
		case <-time.After(time.Duration(s.liveTime) * time.Second):
			// 超过活跃时间了，执行关闭
			s.finalStop()
			return
		}

	}
}

// 关闭链接
func (s *Conn) stop() {
	// 使用上下文方法 关闭  cal()执行后 Start()方法中 ctx.done 就会收到
	s.cal()
}

func (s *Conn) read() {

	defer s.stop()

	for {

		select {
		case <-s.ctx.Done(): // 上下文关闭了，退出方法
			return
		default:
			// 处理连接
			// 封装请求数据
			req := newRequest(s)
			// mqtt 协议，这里接收的数据就是 订阅和发布数据
			err := req.getMqttProto()

			if err != nil {
				// 获取协议错误，直接退出方法
				fmt.Println("err===", err)
				return
			}

			// 更新状态
			s.liveChan <- true

			// 使用协程池
			if s.ofServer.routerMer.workPoolIsOn() {

				s.ofServer.routerMer.sendReqToTaskQueue(req)

			} else {
				// 开启协程处理 路由  获取到协议，根据协议处理数据
				go s.ofServer.routerMer.doRouterFunc(req)
			}

		}

	}

}

func (s *Conn) write() {

	for {

		select {
		case <-s.ctx.Done(): // 上下文关闭了 退出方法
			return
		case data, ok := <-s.writerBuffChan:

			// 有缓冲通道
			if ok {
				//有数据要写给客户端
				if _, err := s.netConn.Write(data); err != nil {
					//fmt.Println("写给客户端数据失败:, ", err, "连接退出")
					return
				}
			} else {
				// 有缓冲写通道关闭了
				break
			}

		default:
			// 将 需要写入的数据

		}
	}

}

func (s *Conn) getClientID() string {

	return s.clientID
}

func (s *Conn) setClientID() error {

	if s.clientID != "" {
		return errors.New("client 已经存在")
	}

	request := newRequest(s)

	// 第一次链接 协议 类型  CONNECT
	p, code, err := request.getCONNECT()

	if err != nil {
		// 直接关闭链接，不发响应
		return err
	}

	if code == 0 {

		// code = 0 可以进行链接 需要接入自定义的链接验证 链接验证仅执行一次
		code = s.ofServer.routerMer.connectVerify(p)

		if code == 0 {
			// 根据协议，在路由管理器中搜索对应的路由对象，执行方法
			s.ofServer.routerMer.doRouterFunc(request)

		}

	}

	// 返回确认消息
	err2 := request.sendCONNACK(code)

	if err2 != nil {
		return err2
	}

	/*
		1，要确认 code = 0
		2，要确认 链接响应没有报错
		才可以加入链接map
	*/

	if code != 0 {
		return errors.New("链接信息错误，不可以链接")
	}

	// 获取clientID 之后要在server 的 链接对象管理器中添加数据
	s.clientID = p.ClientID
	// 获取clientID 之后要在server 的 链接对象管理器中添加数据
	s.ofServer.connMer.addConn(s)
	// 验证遗嘱标识，进行存储
	if p.WillFlag {
		s.ofServer.topicMer.setClientWill(p.ClientID, &proto.Will{
			WillTopic:   p.WillTopic,
			WillMessage: p.WillMessage,
			WillRetain:  p.WillRetain,
			WillQos:     p.WillQos,
		})
	}

	return nil

}

// 获取socket 套接字
func (s *Conn) getTcpConn() *net.TCPConn {
	return s.netConn
}

/*
写通道接收数据
*/
func (s *Conn) sendByte(by []byte) {
	s.writerBuffChan <- by
}

/*
最终关闭，关闭所有资源，关闭 tcp 链接
*/
func (s *Conn) finalStop() {

	if s.isClose == true {
		// 已经关闭不处理
		return
	}

	s.isClose = true

	// 关闭写通道
	close(s.writerBuffChan)
	// 关闭 livechan
	close(s.liveChan)

	s.netConn.Close()

	// 发送遗嘱
	s.ofServer.topicMer.sendClientWill(s.clientID)

	// 删除遗嘱
	s.ofServer.topicMer.removeClientWill(s.clientID)

	// 链接管理器中移出
	s.ofServer.connMer.removeConn(s.clientID)

	zaplog.ZapLogger.Info("【连接关闭】", zap.String("client", s.clientID))

}

/*
设置 属性=》 值
*/

func (s *Conn) setKeyValue(key string, value interface{}) {
	s.kvLock.Lock()
	defer s.kvLock.Unlock()

	s.keyValue[key] = value
}

/*
获取 属性=》 值
*/
func (s *Conn) getKeyValue(key string) (interface{}, error) {

	s.kvLock.RLock()
	defer s.kvLock.RUnlock()

	v, ok := s.keyValue[key]
	if ok {
		return v, nil
	}
	return nil, errors.New("没有此key")
}

/*
移除 属性=》 值
*/
func (s *Conn) removeKeyValue(key string) {

	s.kvLock.Lock()
	defer s.kvLock.Unlock()

	delete(s.keyValue, key)

}
