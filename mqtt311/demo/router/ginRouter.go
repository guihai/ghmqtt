package router

import (
	"github.com/gin-gonic/gin"
	"github.com/guihai/ghmqtt/mqtt311/server"
	"github.com/guihai/ghmqtt/mqtt311/server/types"
)

var MQTTAPI *server.GHapi

func RouterInit(r *gin.Engine) {

	// 404 错误
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, map[string]interface{}{
			"Code": 404,
			"Msg":  "404错误",
		})
	})

	r.POST("", func(c *gin.Context) {
		c.JSON(200, map[string]interface{}{
			"Code": 200,
			"Msg":  "post",
		})
	})
	r.GET("", func(c *gin.Context) {
		c.JSON(200, map[string]interface{}{
			"Code": 200,
			"Msg":  "get",
		})
	})

	MqttServerRouter(r)

}

// 无需jwt 即可进行访问
func MqttServerRouter(r *gin.Engine) {

	group := r.Group("/mqtt/server")

	// 用户进入，用户信息进入换jwt
	group.POST("/info", func(c *gin.Context) {

		c.JSON(200, MQTTAPI.ServerInfo())
	})

	// 发布消息
	group.POST("/publish", func(c *gin.Context) {

		data := &types.PublishMsg{}

		c.BindJSON(data)

		c.JSON(200, MQTTAPI.SendPublish(data))
	})

	// 设置主题 保留消息
	group.POST("/setRetain", func(c *gin.Context) {

		data := &types.PublishMsg{}

		c.BindJSON(data)

		c.JSON(200, MQTTAPI.SetRetainMsg(data))
	})

	// 获取链接列表
	group.POST("/clientList", func(c *gin.Context) {

		c.JSON(200, MQTTAPI.GetConnList())
	})

	// 获取订阅主题列表
	group.POST("/topList", func(c *gin.Context) {

		c.JSON(200, MQTTAPI.GetTopList())
	})
}
