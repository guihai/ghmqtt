package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

/*
链接协议结构体  客户端到服务端
CONNECT = 0x10
*/
type CONNECTProtocol struct {
	// 固定报头
	*Fixed
	// 可变报头
	ProtoNameLen uint16 // 两个字节  默认 4
	ProtoName    string // 根据ProtoNameLen 长度获取的数据
	Version      uint8  // 一个字节 协议级别 版本号  mqtt5.0 版本 0x05
	ConnectFlag  uint8  // 一个字节  链接标志 可以控制有效载荷
	KeepAlive    uint16 // 两个字节 大端解码返回数值

	// 属性 内容 属性标识符后面跟随的就是属性内容
	PropertiesLength      uint32 // 1-4 字节 和 msgLen 相同
	SessionExpiryInterval uint32 // 会话过期间隔	四字节整数
	AuthenticationMethod  string // 认证方法
	//AuthenticationData         []byte            // 认证数据
	AuthenticationData         string            // 认证数据
	RequestProblemInformation  uint8             //请求问题信息
	RequestResponseInformation uint8             // 请求响应信息
	ReceiveMaximum             uint16            //接收最大数量
	TopicAliasMaximum          uint16            // 主题别名最大长度
	UserProperty               map[string]string // 用户属性	字符串对 [k1,v1][k2,v2]
	MaximumPacketSize          uint32            //最大报文长度

	// 有效载荷 根据可变报头参数  ConnectFlag 这里的数据有变化
	/*
		客户标识符(Client Identifier)、
		遗嘱属性(Will Properties)、
		遗嘱主题(Will Topic)、
		遗嘱载荷(Will Payload)、
		用户名(User Name)、
		密码(Password)的顺序出现
	*/

	ClientIDLength uint16 // 两个字节  大端解码返回数值  可以设定设定长度不要超过 200
	ClientID       string // 根据长度 获取数据

	WillProperties         uint32 // 遗嘱属性长度
	PayloadFormatIndicator uint8
	MessageExpiryInterval  uint32
	ContentType            string
	ResponseTopic          string
	//CorrelationData        []byte
	CorrelationData   string
	WillDelayInterval uint32

	WillTopicLength uint16 // 大端解码返回数值
	WillTopic       string

	WillMessageLength uint16
	WillMessage       string

	UserNameLength uint16
	UserName       string

	PasswordLength uint16
	Password       string

	// 以下数据根据 ConnectFlag 解析出使用
	CleanStart bool
	WillFlag   bool
	WillRetain bool
	WillQos    uint8
	// AckCode 生成对应响应使用的AckCode
	AckCode uint8
}

func NewCONNECTProtocol(f *Fixed) *CONNECTProtocol {
	return &CONNECTProtocol{
		Fixed:        f,
		UserProperty: make(map[string]string),

		AckCode: Success, // 默认成功
	}
}
func NewCONNECTProtocolClient(clientID, user, psd string) *CONNECTProtocol {
	p := &CONNECTProtocol{
		Fixed: &Fixed{
			HeaderFlag: CONNECT,
			MsgLen:     0,
			Data:       nil,
		},
		ProtoNameLen:               4, // 默认4
		ProtoName:                  "MQTT",
		Version:                    5,    // 默认5
		ConnectFlag:                0xC2, // 1100 0010
		KeepAlive:                  60,
		PropertiesLength:           0,
		SessionExpiryInterval:      0,
		AuthenticationMethod:       "",
		AuthenticationData:         "",
		RequestProblemInformation:  0,
		RequestResponseInformation: 0,
		ReceiveMaximum:             0,
		TopicAliasMaximum:          0,
		UserProperty:               nil,
		MaximumPacketSize:          0,
		ClientIDLength:             uint16(len(clientID)),
		ClientID:                   clientID,
		WillProperties:             0,
		PayloadFormatIndicator:     0,
		MessageExpiryInterval:      0,
		ContentType:                "",
		ResponseTopic:              "",
		CorrelationData:            "",
		WillDelayInterval:          0,
		WillTopicLength:            0,
		WillTopic:                  "",
		WillMessageLength:          0,
		WillMessage:                "",
		UserNameLength:             uint16(len(user)),
		UserName:                   user,
		PasswordLength:             uint16(len(psd)),
		Password:                   psd,
		CleanStart:                 false,
		WillFlag:                   false,
		WillRetain:                 false,
		WillQos:                    0,
		AckCode:                    0,
	}

	p.Fixed.MsgLen = uint32(11 + 2 + p.ClientIDLength + 2 + p.UserNameLength + 2 + p.PasswordLength)
	return p
}

func (s *CONNECTProtocol) Pack() ([]byte, error) {

	// 固定报头
	by := make([]byte, 1, 14) // 至少14个字节
	by[0] = s.GetHeaderFlag()

	by = append(by, s.msgLenCode(s.GetMsgLen())...)

	// 可变报头
	by = append(by, s.int16ToByBig(s.ProtoNameLen)...)
	by = append(by, []byte(s.ProtoName)...)
	by = append(by, s.Version, s.ConnectFlag)
	by = append(by, s.int16ToByBig(s.KeepAlive)...)
	by = append(by, s.msgLenCode(s.PropertiesLength)...) // 属性默认0

	// clientID
	by = append(by, s.int16ToByBig(s.ClientIDLength)...)
	by = append(by, []byte(s.ClientID)...)

	// user
	by = append(by, s.int16ToByBig(s.UserNameLength)...)
	by = append(by, []byte(s.UserName)...)

	//psd
	by = append(by, s.int16ToByBig(s.PasswordLength)...)
	by = append(by, []byte(s.Password)...)

	return by, nil

}

func (s *CONNECTProtocol) UnPack() error {

	// 检测f 的基本长度  10(基础) + 1（属性长度1） +  3(至少有 clientid)
	if s.Fixed.MsgLen < 14 {
		s.AckCode = Malformed_Packet
		// "数据长度错误"
		return errors.New("数据长度错误")

	}

	if s.Fixed.HeaderFlag != CONNECT {
		s.AckCode = Malformed_Packet
		// "数据长度错误"
		return errors.New("协议类型错误")
	}

	//  移动坐标
	var indx uint16 = 0
	indx = 2 // ProtoNameLen
	daBy := s.Fixed.Data

	binary.Read(bytes.NewBuffer(daBy[:indx]),
		binary.BigEndian, &s.ProtoNameLen)

	if s.ProtoNameLen != 4 { // 默认是4
		s.AckCode = Protocol_Error
		return errors.New("协议错误")
	}

	s.ProtoName = string(daBy[indx:(indx + s.ProtoNameLen)]) // 从 索引[2] 取对应长度值

	if s.ProtoName != "MQTT" { // 默认是MQTT
		s.AckCode = Protocol_Error
		return errors.New("协议错误")
	}

	// 移动索引 2+4 = 6
	indx = indx + s.ProtoNameLen

	s.Version = daBy[indx] // 版本号  5  0x05
	if s.Version != 0x05 { // 默认是 0x05
		s.AckCode = Protocol_Error
		return errors.New("协议错误")
	}

	indx = indx + 1            // 7
	s.ConnectFlag = daBy[indx] // 后面的字节 是 连接标志

	indx = indx + 1 // 8
	// 计算标识符id  大端编码 字节写入数字
	binary.Read(bytes.NewBuffer(daBy[indx:indx+2]),
		binary.BigEndian, &s.KeepAlive)

	indx = indx + 2 // 10
	// 属性长度
	//s.PropertiesLength = daBy[indx]
	//s.PropertiesLength = s.by2Len32(daBy[indx:])
	// 属性长度占据字节数
	var cou uint32 = 1 // 至少是1,最大是4
	s.PropertiesLength, cou = s.by2Len32AndIndex(daBy[indx:])

	indx = indx + uint16(cou) // 11 或者 15

	if s.PropertiesLength > 0 {
		// 有属性值，需要处理
		temp := daBy[indx:(indx + uint16(s.PropertiesLength))]

		var tinx uint32 = 0

		// 移动坐标  最后一位 是 len-1
		for tinx < s.PropertiesLength {

			switch temp[tinx] {
			case SessionEI:
				// 四字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+4]),
					binary.BigEndian, &s.SessionExpiryInterval)
				tinx = tinx + 1 + 4

			case ReceiveMaximum:
				// 双字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
					binary.BigEndian, &s.ReceiveMaximum)
				tinx = tinx + 1 + 2

			case MaximumPS:
				//四字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+4]),
					binary.BigEndian, &s.MaximumPacketSize)
				tinx = tinx + 1 + 4

			case TopicAM:
				//双字节整数
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
					binary.BigEndian, &s.TopicAliasMaximum)
				tinx = tinx + 1 + 2

			case UserProperty:
				// 用户属性
				// len 两个字节
				var keyLen uint16
				binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
					binary.BigEndian, &keyLen)

				tinx += 1 + 2
				key := string(temp[tinx : tinx+uint32(keyLen)])

				tinx += uint32(keyLen)

				var valLen uint16
				binary.Read(bytes.NewBuffer(temp[tinx:tinx+2]),
					binary.BigEndian, &valLen)

				tinx += 2

				val := string(temp[tinx : tinx+uint32(valLen)])

				tinx += uint32(valLen)

				s.UserProperty[key] = val

			case AuthenticationM:
				//
				fmt.Println("AuthenticationM")
			case AuthenticationD:
				//
				fmt.Println("AuthenticationD")
			case RequestPI:
				//
				fmt.Println("RequestPI")
			case RequestRI:
				//
				fmt.Println("RequestRI")

			default:
				// 匹配不到 结束循环
				break

			}

		}

		// 坐标移动
		indx = indx + uint16(s.PropertiesLength)

	}

	// 没有属性值
	// 拆解 有效载荷
	daByp := daBy[indx:] // 获取有效载荷的后续数据

	if len(daByp) < 3 {
		s.AckCode = ClientInotv
		return errors.New("第一次链接必须有 client 必须有值")
	}

	s.ClientIDLength, s.ClientID, _ = s.by2LenNameBE(daByp)

	if s.ClientIDLength < 1 || s.ClientID == "" {
		s.AckCode = ClientInotv
		return errors.New("第一次链接必须有 client 必须有值")
	}

	daByp2 := daByp[2+s.ClientIDLength:] // 获取后续数据
	// 数字变成 8位二级制
	bs := fmt.Sprintf("%08b", s.ConnectFlag)

	// 根据  ConnectFlag 获取其他数据
	// 48 == 0 ，49 ==1
	if bs[7] != 48 {
		// CONNECT 控制报文的保留标志位（第 0 位）是否为 0，如果不为 0 必须断开客户端 连接
		s.AckCode = Malformed_Packet
		return errors.New("CONNECT 控制报文的保留标志位（第 0 位）必须是0")
	}

	// CleanStart=1  1新会话 ，0 关联之前的回话
	if bs[6] == 49 { // 0
		// 1 不用处理 0 要保留回话，在业务端处理
		s.CleanStart = true
	}

	// Will Flag  遗嘱标志  遗嘱主题，遗嘱消息，当客户端断开链接时，要将客户端的遗嘱 向订阅遗嘱主题的客户端发布
	// 让其他客户端 获取某个 客户端下线
	if bs[5] == 49 { // 1 需要 在有效载荷中获取 遗嘱数据
		// 获取 遗嘱有效数据 遗嘱主题和遗嘱消息 至少6个字节 + 遗嘱属性总长度1
		if len(daByp2) < 7 {
			// 不存在遗嘱
			s.AckCode = Malformed_Packet
			return errors.New("不存在遗嘱")

		}

		var idx uint32 = 1 // 至少是1
		s.WillProperties, idx = s.by2Len32AndIndex(daByp2)
		// 获取遗嘱属性
		if s.WillProperties > 0 {
			// 遗嘱有属性
			temp := daByp2[idx : idx+s.WillProperties]
			var tinx uint32 = 0
			// 移动坐标  最后一位 是 len-1
			for tinx < s.WillProperties {

				switch temp[tinx] {
				case PayloadFI:
					// 一个字节
					s.PayloadFormatIndicator = daByp2[tinx+1]
					// 移动索引
					tinx += tinx + 1 + 1

				case MessageEI:
					// 四字节整数
					binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+4]),
						binary.BigEndian, &s.MessageExpiryInterval)
					tinx = tinx + 1 + 4

				case ContentType:
					//先获取长度
					var ctLen uint16 = 0
					// 再获取数据
					binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
						binary.BigEndian, &ctLen)

					tinx += 1 + 2

					s.ContentType = string(temp[tinx : tinx+uint32(ctLen)])

					tinx += uint32(ctLen)

				case ResponseTopic:
					//先获取长度
					var ctLen uint16 = 0
					//双字节整数
					binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
						binary.BigEndian, &ctLen)
					tinx = tinx + 1 + 2

					s.ResponseTopic = string(temp[tinx : tinx+uint32(ctLen)])

					tinx += uint32(ctLen)

				case CorrelationData:
					//先获取长度
					var ctLen uint16 = 0
					//双字节整数
					binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+2]),
						binary.BigEndian, &ctLen)
					tinx = tinx + 1 + 2

					s.CorrelationData = string(temp[tinx : tinx+uint32(ctLen)])

					tinx += uint32(ctLen)

				case WillDI:
					// 四字节整型
					binary.Read(bytes.NewBuffer(temp[tinx+1:tinx+1+4]),
						binary.BigEndian, &s.WillDelayInterval)
					tinx = tinx + 1 + 4

				default:
					// 匹配不到 结束循环
					break

				}

			}
			// 截取数据
			daByp2 = daByp2[idx+s.WillProperties:]
		}

		s.WillTopicLength, s.WillTopic, _ = s.by2LenNameBE(daByp2)
		if s.WillTopicLength < 1 || s.WillTopic == "" {
			s.AckCode = Topic_Name_invalid
			return errors.New("遗嘱名无效")
		}

		// 获取遗嘱主题成功，获取遗嘱消息
		s.WillMessageLength, s.WillMessage, _ = s.by2LenNameBE(daByp2[2+s.WillTopicLength:])
		if s.WillMessageLength < 1 || s.WillMessage == "" {
			s.AckCode = Topic_Name_invalid
			return errors.New("遗嘱名无效")
		}

		s.WillFlag = true

		// 需要处理遗嘱相关标识
		if bs[2] == 49 {
			s.WillRetain = true

		}

		// 01
		if bs[3] == 48 && bs[4] == 49 {
			s.WillQos = QoS1
		}

		// 10
		if bs[3] == 49 && bs[4] == 48 {
			s.WillQos = QoS2
		}

		// 以上获取遗嘱完成  修改剩余 字节的长度  必须修改
		daByp2 = daByp2[2+s.WillTopicLength+2+s.WillMessageLength:]

	}

	// User Name Flag
	if bs[0] == 49 {
		if len(daByp2) < 3 {
			// 不存在用户名
			s.AckCode = BadUNorP
			return errors.New("不存在用户名")

		}
		s.UserNameLength, s.UserName, _ = s.by2LenNameBE(daByp2)

		if s.UserNameLength < 1 || s.UserName == "" {
			// 不存在用户名
			s.AckCode = BadUNorP
			return errors.New("不存在用户名")
		}

		// 用户名获取完成 必须修改剩余字节
		daByp2 = daByp2[2+s.UserNameLength:]

	}

	// Password Flag (1) 密码标志
	if bs[1] == 49 {
		if len(daByp2) < 3 {
			// 不存在密码
			s.AckCode = BadUNorP
			return errors.New("不存在密码")
		}
		s.PasswordLength, s.Password, _ = s.by2LenNameBE(daByp2)

		if s.PasswordLength < 1 || s.Password == "" {
			// 不存在密码
			s.AckCode = BadUNorP
			return errors.New("不存在密码")
		}

	}

	// 返回成功
	return nil
}

func (s *CONNECTProtocol) GetAckCode() uint8 {
	return s.AckCode
}
