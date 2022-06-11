package server

import (
	"github.com/guihai/ghmqtt/mqtt311/proto"
	"sync"
)

/*
主题管理对象
在配置中选择使用
*/
type TopicManager struct {
	// 订阅map  key 是主题名称  值是 map( client是key  值空结构体不占内存)
	//SubMap map[string]*TSet
	subMapM map[string]map[string]struct{}

	// 主题信息通道 top 是key  值是通道 ，通道内数据是 byte
	topicMsg map[string]chan []byte
	// 锁
	subLock sync.RWMutex

	// 所属服务
	ofServer *Server

	// 保留消息map
	retainMsg map[string][]byte
	// 保留消息锁
	retainLock sync.RWMutex

	// Qos2 需要存储消息  key 是 标识符 Identifier 内容空结构体,不占内存，只为了存key
	qos2ID map[uint16]struct{}
	// 锁
	qos2Lock sync.RWMutex

	// topmanger
	tm *TopicWork

	// 链接的遗嘱消息  key 是 client 遗嘱是结构体，每个客户只有一个will
	clientWill map[string]*proto.Will
	// 锁
	clientWillLock sync.RWMutex
}

func newTopicManager(ser *Server) *TopicManager {

	t := &TopicManager{
		subMapM:  make(map[string]map[string]struct{}),
		ofServer: ser,
		topicMsg: make(map[string]chan []byte),

		// 保留消息map
		retainMsg: make(map[string][]byte),

		// 保留标识符map
		qos2ID: make(map[uint16]struct{}),

		// 客户端的遗嘱消息
		clientWill: make(map[string]*proto.Will),
	}

	// 开启主题协程池
	t.tm = newTopicWork(t)
	t.tm.startWorkerPool()

	return t
}

// 使用协程池 减少每个发布的主题都要开启协程 的协程数量
func (s *TopicManager) msgInPool(p *proto.PUBLISHProtocol) {
	s.tm.sendReqToTaskQueue(p)
}

/*
订阅主题
1，不校验是否存在，map 保证数据唯一
2,如果没有用户，client 为空
*/

func (s *TopicManager) subTopic(top, client string) {

	if len(s.subMapM[top]) < 1 {
		// 这个主题还 没有初始化
		s.subMapM[top] = make(map[string]struct{})
	}

	s.subLock.Lock()
	s.subMapM[top][client] = struct{}{}
	s.subLock.Unlock()

	// 用户订阅主题后，可以先发送 保留信息
	s.sendRetainMsg(top, client)

}

/*
取消订阅
*/
func (s *TopicManager) unSubTopic(top, client string) {
	if len(s.subMapM[top]) < 1 {
		// 这个主题不存在
		return
	}

	s.subLock.Lock()
	delete(s.subMapM[top], client)

	if len(s.subMapM[top]) < 1 {
		delete(s.subMapM, top)
	}
	s.subLock.Unlock()

}

/*
查询主题的订阅用户 返回用户 列表
*/
func (s *TopicManager) getTopSubList(top string) []string {

	tlen := len(s.subMapM[top])
	if tlen < 1 {
		// 不存在主题
		return []string{}
	}
	// 存在主题
	sl := make([]string, tlen)

	s.subLock.RLock()

	for cli, _ := range s.subMapM[top] {
		sl = append(sl, cli)
	}

	s.subLock.RUnlock()

	return sl
}

/*
获取所有主题列表
*/
func (s *TopicManager) getTopList() []string {

	tlen := len(s.subMapM)
	if tlen < 1 {
		// 不存在主题
		return []string{}
	}
	// 存在主题
	sl := make([]string, 0, tlen)

	s.subLock.RLock()

	for top, _ := range s.subMapM {
		sl = append(sl, top)
	}

	s.subLock.RUnlock()

	return sl
}

/*
设置保留消息
每个主题只保存一条保留消息，所以可以直接赋值，如果byte 为空，就是删除
*/
func (s *TopicManager) setRetainMsg(top string, by []byte) {

	if len(by) < 1 {
		s.retainLock.Lock()
		delete(s.retainMsg, top)
		s.retainLock.Unlock()
	}
	// 直接赋值
	s.retainLock.Lock()
	s.retainMsg[top] = by
	s.retainLock.Unlock()

	//fmt.Println("存储保留消息成功，", top)
	//fmt.Println("保留消息个数", len(s.RetainMsg))

}

/*
获取保留信息
*/
func (s *TopicManager) getRetainMsg(top string) ([]byte, bool) {
	// todo 通配符保留信息

	s.retainLock.RLock()
	by, ok := s.retainMsg[top]
	s.retainLock.RUnlock()

	if !ok {
		return []byte{}, ok
	}

	return by, ok
}

/*
发送保留消息
*/
func (s *TopicManager) sendRetainMsg(top string, client string) {

	bys, ok := s.getRetainMsg(top)

	if !ok || len(bys) < 1 {
		return
	}

	dp := newMqttDataPack()
	// todo 创建发布协议   按照 48 创建
	p := &proto.PUBLISHProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PUBLISH,
			MsgLen:     0,
			Data:       nil,
		},
		TopicNameLength:  uint16(len(top)),
		TopicName:        top,
		PacketIdentifier: [2]byte{},
		Payload:          bys,
	}

	// 长度标识2 个
	p.MsgLen = 1 + 1 + uint32(p.TopicNameLength) + uint32(len(bys))

	by2, err := dp.packPUBLISH(p)

	// 发送数据
	con, err := s.ofServer.connMer.getConn(client)

	if err != nil {
		return
	}

	con.sendByte(by2)

}

/*
保存 Qos2 标识符
不需要校验是否已存在，直接添加
*/
func (s *TopicManager) setQos2ID(id uint16) {
	// 直接赋值
	s.qos2Lock.Lock()
	s.qos2ID[id] = struct{}{}
	s.qos2Lock.Unlock()
}

/*
查找 Qos2ID 对应的key 存在就返回true
*/
func (s *TopicManager) getQos2ID(id uint16) bool {
	s.qos2Lock.RLock()
	_, ok := s.qos2ID[id]
	s.qos2Lock.RUnlock()

	return ok
}

/*
完成Qos2 全部流程 移出key
*/
func (s *TopicManager) removeQos2ID(id uint16) {
	// 直接删除
	s.qos2Lock.Lock()
	delete(s.qos2ID, id)
	s.qos2Lock.Unlock()

}

/*
设置遗嘱消息
*/
func (s *TopicManager) setClientWill(client string, p *proto.Will) {
	// 直接赋值
	s.clientWillLock.Lock()
	s.clientWill[client] = p
	s.clientWillLock.Unlock()
}

/*
获取遗嘱消息
*/
func (s *TopicManager) getClientWill(client string) (*proto.Will, bool) {
	// 直接赋值
	s.clientWillLock.RLock()

	val, ok := s.clientWill[client]

	s.clientWillLock.RUnlock()

	if !ok {
		return nil, ok
	}

	return val, ok
}

/*
客户端正常 关闭链接 ,遗嘱遗嘱信息
*/
func (s *TopicManager) removeClientWill(client string) {
	// 直接删除
	s.clientWillLock.Lock()
	delete(s.clientWill, client)
	s.clientWillLock.Unlock()

}

/*
建立遗嘱的客户端，关闭
发送其遗嘱 信息给订阅者客户端
*/
func (s *TopicManager) sendClientWill(client string) {

	will, ok := s.getClientWill(client)

	if !ok {
		return
	}

	dp := newMqttDataPack()
	// todo 创建发布协议   按照 48 创建
	p := &proto.PUBLISHProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PUBLISH,
			MsgLen:     0,
			Data:       nil,
		},
		TopicNameLength:  uint16(len(will.WillTopic)),
		TopicName:        will.WillTopic,
		PacketIdentifier: [2]byte{},
		Payload:          []byte(will.WillMessage),
	}

	// 长度标识2 个
	p.MsgLen = 1 + 1 + uint32(p.TopicNameLength) + uint32(len(p.Payload))

	by2, err := dp.packPUBLISH(p)

	if err != nil {
		return
	}

	// 发送数据
	for scli, _ := range s.subMapM[will.WillTopic] {

		con, err := s.ofServer.connMer.getConn(scli)

		if err != nil {
			continue
		}
		con.sendByte(by2)
	}

}

/*
清理所有资源
*/
func (s *TopicManager) stop() {
	// 所有通道关闭
	for _, c := range s.topicMsg {
		close(c)
	}
}
