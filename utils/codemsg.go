package utils

// 返回code和msg的标准库
const (
	StatusOK = 200
	//StatusMovedPermanently = 301   // 永久定向，少用
	StatusFound = 302 // 临时定向
	//StatusNotFound = 404

	RECODE_OK = 200

	RECODE_NOPOWER = 2001 // 没有权限
	RECODE_CSRFERR = 2002 // Csrf错误

	RECODE_NOJWT  = 2003 // 请求头中Jwt为空
	RECODE_JWTERR = 2005 // 无效的jwt

	RECODE_UNKNOWERR = 4001 // 原始错误
	RECODE_BINDERR   = 4002 // 参数绑定错误
	RECODE_PARAMERR  = 4003 // 参数错误
	RECODE_IOERR     = 4004 // 写入错误
	RECODE_DBERR     = 4005 // 数据库查询错误
	RECODE_RPCERR    = 4006 // RPC调用错误
	RECODE_SMSERR    = 4007 // "短信失败"

	RECODE_NEEDPAY = 4008 // "需要付费开通"

	RECODE_NODATA = 4009 // 数据不存在

	RECODE_REDISINIT = 4010 // redis初始化

	RECODE_FILEERR = 4011 // 文件错误

	RECODE_REPEAT = 4012 // 重复提交

	RECODE_HASDATA = 4013 // 数据已存在

)

var recodeText = map[int]string{
	RECODE_OK: "OK",

	RECODE_NOPOWER: "没有权限",

	RECODE_CSRFERR: "Csrf错误",

	RECODE_NOJWT: "Jwt为空",

	RECODE_JWTERR: " 无效的Jwt",

	RECODE_DBERR: "数据库查询错误",

	RECODE_PARAMERR: "参数错误",

	RECODE_IOERR: "文件读写错误",

	RECODE_UNKNOWERR: "未知错误",

	RECODE_BINDERR: "数据绑定错误",

	RECODE_RPCERR: "RPC调用错误",

	RECODE_SMSERR: "短信失败",

	RECODE_NEEDPAY: "需要付费开通",

	RECODE_NODATA: "数据不存在",

	RECODE_REDISINIT: "redis初始化",

	RECODE_FILEERR: "文件错误",

	RECODE_REPEAT: "重复提交",

	RECODE_HASDATA: "数据已存在",
}

//函数  根据key来获取value

func MsgText(code int) string {
	str, ok := recodeText[code]
	if ok {
		return str
	}
	return recodeText[RECODE_UNKNOWERR]
}
