package server

import (
	"errors"
	"github.com/guihai/ghmqtt/utils/zaplog"
	"sync"
)

type ConnManager struct {
	// 连接对象Map  clientid 是key
	connMap map[string]*Conn

	// map 的读写锁
	mapLock sync.RWMutex
}

func newConnManager() *ConnManager {
	return &ConnManager{
		connMap: make(map[string]*Conn),
		//mapLock: sync.RWMutex{},
	}
}

/*
添加链接
1.如果验证是否已存在，新的链接不能替换原来的，实际中可能需要新的来替换旧的，所以不要验证是否已存在，
*/

func (s *ConnManager) addConn(conn *Conn) bool {
	// 上锁
	s.mapLock.Lock()
	// 添加数据 不验证是否已存在，用新链接替换旧的链接
	s.connMap[conn.clientID] = conn
	// 解锁
	s.mapLock.Unlock()

	return true
}

/*
获取 对象map 长度
*/
func (s *ConnManager) getLen() int {

	// 上读锁
	s.mapLock.RLock()
	l := len(s.connMap)
	// 解读锁
	s.mapLock.RUnlock()
	return l

}

/*
根据 clientid 获取conn
*/
func (s *ConnManager) getConn(clientID string) (*Conn, error) {
	// 上读锁
	s.mapLock.RLock()
	c, ok := s.connMap[clientID]
	// 解读锁
	s.mapLock.RUnlock()

	if !ok {
		return nil, errors.New("链接不存在")
	}
	return c, nil

}

/*
在线链接列表
*/
func (s *ConnManager) getConnList() []string {
	// 设定长度，容量
	list := make([]string, 0, s.getLen())
	// 上读锁
	s.mapLock.RLock()

	for client, _ := range s.connMap {
		list = append(list, client)
	}
	// 解读锁
	s.mapLock.RUnlock()
	return list

}

/*
移出 对象
1,无需验证是否存在，不存在删除也不报错
2，删除对象
*/
func (s *ConnManager) removeConn(cilentID string) {

	s.mapLock.Lock()
	delete(s.connMap, cilentID)
	s.mapLock.Unlock()

	return

}

/*
清空map
服务关闭，删除所有conn
*/

func (s *ConnManager) clearConn() {

	s.mapLock.Lock()
	for key, conn := range s.connMap {

		// 停止链接服务
		conn.stop()
		// 删除key
		delete(s.connMap, key)

	}
	zaplog.ZapLogger.Info("【清空所有链接】")
	s.mapLock.Unlock()

	return
}
