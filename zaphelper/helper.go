package zaphelper

import (
	"errors"
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	HelperConfig struct {
		LoggerName string
		Filename   string
		MaxSize    int
		MaxAge     int
		MaxBackups int
		LocalTime  bool
		Compress   bool
	}
)

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
	if hc.Filename == "" {
		panic(errors.New("no file path provided"))
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

	logger := zap.New(core, buildOptionsFromConfig(conf, writer)...)
	if hc.LoggerName != "" {
		logger = logger.Named(hc.LoggerName)
	}
	if len(opts) > 0 {
		logger = logger.WithOptions(opts...)
	}
	return logger
}
