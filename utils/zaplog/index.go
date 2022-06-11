package zaplog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

type LogConfig struct {
	Filename   string `json:"Filename" yaml:"Filename"`
	MaxSize    int    `json:"MaxSize" yaml:"MaxSize"`
	MaxAge     int    `json:"MaxAge" yaml:"MaxAge"`
	MaxBackups int    `json:"MaxBackups" yaml:"MaxBackups"`
	//LocalTime  bool   `json:"LocalTime" yaml:"LocalTime"`
	Compress bool `json:"Compress" yaml:"Compress"`

	// 日志级别
	Level zapcore.Level `json:"Level" yaml:"Level"` // - 1 =》 5

	//是否空值台输出
	StdoutFLag bool `json:"StdoutFLag" yaml:"StdoutFLag"`
}

var ZapLogger *zap.Logger

func InitLogger(cfg *LogConfig) {
	hook := lumberjack.Logger{
		Filename:   cfg.Filename,   // 日志文件路径
		MaxSize:    cfg.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: cfg.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     cfg.MaxAge,     // 文件最多保存多少天
		Compress:   cfg.Compress,   // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	//atomicLevel.SetLevel(zap.InfoLevel)
	atomicLevel.SetLevel(cfg.Level)

	var wType zapcore.WriteSyncer
	// 设置输出类型
	if cfg.StdoutFLag {
		// 控制台输出 打印到控制台和文件
		wType = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		// 打印到文件
		wType = zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		wType,                                 // 打印到控制台和文件
		atomicLevel,                           // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段
	//filed := zap.Fields(zap.String("serviceName", "serviceName"))
	// 构造日志
	ZapLogger = zap.New(core, caller, development)

	ZapLogger.Info("log 初始化成功")
	//logger.Info("无法获取网址",
	//	zap.String("url", "http://www.baidu.com"),
	//	zap.Int("attempt", 3),
	//	zap.Duration("backoff", time.Second))

}
