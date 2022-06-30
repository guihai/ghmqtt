package server

import (
	"fmt"
	"github.com/guihai/ghmqtt/mqtt5/proto"
	"github.com/guihai/ghmqtt/utils"
	"github.com/guihai/ghmqtt/utils/zaplog"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"time"
)

type TopicWork struct {

	// 创建协程池 工作单位+任务队列
	// 工作单位数量
	workPoolSize uint32
	// 任务队列 管道 切面，数据就是请求
	taskQueue []chan *proto.PUBLISHProtocol

	// WorkPoolIsOn
	poolOn bool

	// 打包工具
	dp *MqttDataPack

	// 所属的topic对象
	ofTopic *TopicManager
}

func newTopicWork(top *TopicManager) *TopicWork {

	r := &TopicWork{
		// 创建协程池
		workPoolSize: utils.GO.WorkPoolSize,
		//一个worker对应一个queue
		taskQueue: make([]chan *proto.PUBLISHProtocol, utils.GO.WorkPoolSize),

		poolOn: false, // 协程池未启动

		// 打包工具
		dp: newMqttDataPack(),

		// 主题对象
		ofTopic: top,
	}

	return r

}

/*
协程池是否开启
*/
func (s *TopicWork) workPoolIsOn() bool {
	return s.poolOn
}

/*
初始化协程池，服务已开启就要启动
*/
func (s *TopicWork) startWorkerPool() {

	if s.poolOn {
		// 已经开启就 不执行了
		return
	}
	for i := 0; i < int(s.workPoolSize); i++ {

		// 初始化任务队列的管道
		s.taskQueue[i] = make(chan *proto.PUBLISHProtocol, utils.GO.TaskQueueMaxSize)

		// 启动一个工作单位
		go s.startWorker(i, s.taskQueue[i])
	}

	s.poolOn = true
}

/*
开启工作单位
*/
func (s *TopicWork) startWorker(i int, pubs chan *proto.PUBLISHProtocol) {
	zaplog.ZapLogger.Info("【PUBLISH协程池】 创建工作者", zap.Int("编号", i))
	//不断的等待队列中的消息,然后进行路由处理
	for {
		select {
		case re := <-pubs:
			s.sendPub(re)
		}
	}

}

//将消息交给TaskQueue,由worker进行处理
func (s *TopicWork) sendReqToTaskQueue(pub *proto.PUBLISHProtocol) {
	// 采用 链接id 取余数 放入对应的消息队列
	cid := pub.TopicName
	//pid := request.Proto.GetHeaderFlag()

	rand.Seed(time.Now().Unix())
	wid := (dHash(cid) + rand.Uint32()) % s.workPoolSize

	fmt.Println("添加 pub topname =", cid, "到 workerID=", wid)

	s.taskQueue[wid] <- pub
}

/*
发布 消息到用户
*/
func (s *TopicWork) sendPub(re *proto.PUBLISHProtocol) {

	// 根据获取的消息协议，创造新的消息协议
	// 按照48 创建 不需要标识符 消息长度
	p := &proto.PUBLISHProtocol{
		Fixed: &proto.Fixed{
			HeaderFlag: proto.PUBLISH,
			MsgLen:     0,
			Data:       nil,
		},
		TopicNameLength:  re.TopicNameLength,
		TopicName:        re.TopicName,
		PacketIdentifier: [2]byte{},
		PropertiesLength: 0, // 属性0
		Payload:          re.Payload,
	}

	// 长度标识2 个 属性 1个
	p.MsgLen = 2 + 1 + uint32(p.TopicNameLength) + uint32(len(p.Payload))

	by2, err := p.Pack()

	if err != nil {
		return
	}

	// 通配符匹配 获取最终要发布的 client
	s.matchSend(re.TopicName, by2)
}

func (s *TopicWork) matchSend(topic string, data []byte) {
	tol := MatchTopic(topic)
	if len(tol) < 1 {
		return // 不发布
	}

	// 链接map
	var conMap = make(map[*Conn]struct{})

	for _, stop := range tol {

		for client, _ := range s.ofTopic.subMapM[stop] {

			con, err := s.ofTopic.ofServer.connMer.getConn(client)

			if err != nil {
				continue
			}

			// 检查是否出现过
			_, ok := conMap[con]
			if ok {
				// 有值，已经存入了
				continue
			}
			// 不存在值
			// 发布
			con.sendByte(data)
			// 添加到map
			conMap[con] = struct{}{}
		}
	}

}

/*
通配符方法 参考 // https://blog.csdn.net/qq_41257365/article/details/115499403
根据主题，获得 通配符后的 多个主题列表
在根据主题列表，在订阅用户中获取订阅者，每个订阅者只获取一次
*/
func MatchTopic(topic string) []string {
	if len(topic) == 1 && (topic == "/" || topic == "#") {
		return []string{}
	}
	tp := strings.Split(topic, "/")
	size := strings.Count(topic, "/")
	ret := make([]string, 0, 1<<(size+1)-1)
	for i := 0; i < 3; i++ {
		dfs(tp, i, 1, tp[0], &ret)
	}
	return ret
}
func dfs(tps []string, index, level int, tempTp string, ret *[]string) {
	if index == 0 {
		*ret = append(*ret, tempTp+"/#")
		return
	} else if index == 1 {
		tempTp = tempTp + "/+"
	} else {
		tempTp = tempTp + "/" + tps[level]
	}
	level++
	if level == len(tps) {
		*ret = append(*ret, tempTp)
		return
	}
	for i := 0; i < 3; i++ {
		dfs(tps, i, level, tempTp, ret)
	}
}
