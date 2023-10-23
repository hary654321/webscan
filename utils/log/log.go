package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"webscan/config"
)

type LoggerInterface interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})

	Info(args ...interface{})
	Infof(template string, args ...interface{})

	Warn(args ...interface{})
	Warnf(template string, args ...interface{})

	Error(args ...interface{})
	Errorf(template string, args ...interface{})
}

type Logger struct {
	Logger *zap.SugaredLogger
}

func NewLogger(configureInterface config.ConfigureInterface) *zap.SugaredLogger {
	conf := configureInterface.GetLogConfig()
	// 设置日志级别
	var level zapcore.Level
	switch conf.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// 定义日志切割配置
	hook := lumberjack.Logger{
		Filename:   path.Join(conf.DataDir, conf.FileName), // 日志文件的位置
		MaxSize:    conf.MaxSize,                           // 在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxBackups: conf.MaxBackups,                        // 保留旧文件的最大个数
		Compress:   conf.Compress,                          // 是否压缩 disabled by default
	}
	if conf.MaxAge > 0 {
		hook.MaxAge = conf.MaxAge // days
	}

	// 判断是否控制台输出日志
	var syncer zapcore.WriteSyncer
	if conf.LogInConsole {
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		syncer = zapcore.AddSync(&hook)
	}

	// 配置日志编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	if conf.JsonFormat {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// 配置日志输出格式
	var encoder zapcore.Encoder
	if conf.JsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置日志核心
	core := zapcore.NewCore(
		encoder,
		syncer,
		level,
	)

	// 配置日志选项
	options := []zap.Option{
		zap.AddCaller(), // 添加代码行号
	}

	if conf.LogInConsole {
		options = append(options, zap.Development()) // 输出到控制台
	}

	// 创建 Logger
	logger := zap.New(core, options...)

	// 构建 Sugared Logger（带有 Sugar 方法的 Logger）
	sugaredLogger := logger.Sugar()

	return sugaredLogger
}

func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.Logger.Debugf(template, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.Logger.Infof(template, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.Logger.Warnf(template, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.Logger.Errorf(template, args...)
}
