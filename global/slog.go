package global

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
	"time"
)

var SystemLog *zap.SugaredLogger

type LogConfig struct {
	Level      string `json:"level"`       // 日志级别: debug, info, warn, error, dpanic, panic, fatal
	FilePath   string `json:"filePath"`    // 日志文件路径
	MaxSize    int    `json:"max_size"`    // 每个日志文件的最大尺寸(单位：MB)
	MaxBackups int    `json:"max_backups"` // 日志文件最多保存多少个备份
	MaxAge     int    `json:"max_age"`     // 文件最多保存多少天
	Compress   bool   `json:"compress"`    // 是否压缩
}

func flushSystemLog() {
	SystemLog = createLog("sentinels")
}

func createLog(name string) *zap.SugaredLogger {
	// 文件核心（带轮转）
	fileCore := newFileCore(zapcore.DebugLevel, name)
	cores := []zapcore.Core{fileCore, newConsoleCore(zapcore.DebugLevel)}
	// 创建tee核心
	core := zapcore.NewTee(cores...)
	// 创建logger，添加调用者信息和堆栈跟踪
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel), // 错误级别以上添加堆栈跟踪
	)
	// 创建SugaredLogger
	return logger.Sugar()
}

func newConsoleCore(level zapcore.LevelEnabler) zapcore.Core {
	consoleEncoder := getConsoleEncoder()
	return zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
}

func getConsoleEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func newFileCore(level zapcore.LevelEnabler, name string) zapcore.Core {
	// 配置lumberjack日志轮转
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path.Join(logPath, name+".slog"),
		MaxSize:    20,    // 单位：MB
		MaxBackups: 7,     // 最大备份数量
		MaxAge:     7,     // 单位：天
		Compress:   false, // 是否压缩
		LocalTime:  true,  // 使用本地时间
	}
	// 文件编码器 - 使用控制台格式
	fileEncoder := getCommonEncoder()
	return zapcore.NewCore(fileEncoder, zapcore.AddSync(lumberJackLogger), level)
}

func getCommonEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}
