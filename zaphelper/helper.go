package zaphelper

import (
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type HelperConfig struct {
	// LoggerName 生成的Logger名称
	LoggerName string

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.
	Filename string

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool
}

func buildOptionsFromConfig(cfg zap.Config, writer zapcore.WriteSyncer) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(writer)}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	stackLevel := zap.ErrorLevel
	if cfg.Development {
		stackLevel = zap.WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, cfg.Sampling.Initial, cfg.Sampling.Thereafter)
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]zap.Field, 0, len(cfg.InitialFields))
		keys := make([]string, 0, len(cfg.InitialFields))
		for k := range cfg.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, cfg.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	return opts
}

// BuildRotateLogger 基于zap.Config以及HelperConfig构建日志
func BuildRotateLogger(conf zap.Config, hc HelperConfig, opts ...zap.Option) *zap.Logger {
	var logger *zap.Logger
	var err error

	if hc.Filename == "" { // 如果不传入文件路径，则视为输出到控制台，不会读取 zap.Config.OutputPaths中的路径
		logger, err = conf.Build(opts...)
		if err != nil {
			panic(err)
		}
		logger = logger.Named(hc.LoggerName)
		return logger
	}

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   hc.Filename,
		MaxSize:    hc.MaxSize,
		MaxAge:     hc.MaxAge,
		MaxBackups: hc.MaxBackups,
		LocalTime:  hc.LocalTime,
		Compress:   hc.Compress,
	})

	var core zapcore.Core
	if conf.Encoding == "json" {
		core = zapcore.NewCore(zapcore.NewJSONEncoder(conf.EncoderConfig), writer, conf.Level)
	} else {
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(conf.EncoderConfig), writer, conf.Level)
	}

	logger = zap.New(core, buildOptionsFromConfig(conf, writer)...)
	if hc.LoggerName != "" {
		logger = logger.Named(hc.LoggerName)
	}
	if len(opts) > 0 {
		logger = logger.WithOptions(opts...)
	}
	return logger
}
