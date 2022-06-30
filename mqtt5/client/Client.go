package client

import (
	"context"
	"fmt"
	"github.com/guihai/ghmqtt/mqtt5/proto"
	"net"
)

type Client struct {
	ip   string // 监听ip 默认 0.0.0.0
	port uint16 // 监听端口 默认  1883
	tcp  string // 传输协议 默认 tcp4

	// 链接对象
	conn net.Conn

	// 退出信号
	ctx context.Context
	cal context.CancelFunc

	ClientID string
	Username string
	Password string

	// 写入数据的 管道
	writeBufChan chan []byte
}

func NewClient(ip string, port uint16, clientid, user, psd string) *Client {

	c := &Client{
		ip:   ip,
		port: port,

		conn: nil,

		ClientID: clientid,
		Username: user,
		Password: psd,

		writeBufChan: make(chan []byte, 1024),
	}

	c.ctx, c.cal = context.WithCancel(context.Background())
	return c
}

func (s *Client) Run() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))

	if err != nil {
		fmt.Println("连接服务器失败", err)
		return
	}

	fmt.Println("链接服务器成功", conn.RemoteAddr())
	s.conn = conn

	// 先要进行链接协议
	s.sendCONNECT()

	// 然后是正常链接

	// 开启协程读写
	go s.Read()
	//
	go s.Write()

	select {
	case <-s.ctx.Done():
		//退出阻塞 结束程序
		s.stop()
		fmt.Println("失去服务器，连接关闭")

		return

	}

}

func (s *Client) Read() {

	defer s.stop()

	for {

		select {
		case <-s.ctx.Done(): // 上下文关闭了，退出方法
			return
		default:
			// 处理连接
			// 封装请求数据
			msg := NewMessage(s.conn)
			// mqtt 协议，这里接收的数据就是 订阅和发布数据
			err := msg.getMqttProto()

			if err != nil {
				// 获取协议错误，直接退出方法
				fmt.Println("err===", err)
				s.cal() // 结束所有程序
				return
			}
			// 处理数据后返回写数据

		}

	}

}

func (s *Client) Write() {

	for bytes := range s.writeBufChan {
		s.conn.Write(bytes)
	}

}

func (s *Client) stop() {
	// todo 断开链接协议

	s.conn.Close()
}

func (s *Client) sendCONNECT() {

	// 创建链接协议
	p := proto.NewCONNECTProtocolClient(s.ClientID, s.Username, s.Password)

	by, err := p.Pack()
	if err != nil {
		// 关闭链接
		s.cal()
		return
	}
	// 发送by
	s.conn.Write(by)
}
