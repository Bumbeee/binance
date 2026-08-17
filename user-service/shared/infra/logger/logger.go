package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerStruct struct {
	Level             string
	Encoding          string
	DisableCaller     bool
	DisableStacktrace bool

	FilePath       string
	FileMaxSizeMB  int
	FileMaxBackups int
	FileMaxAgeDays int
}

func New(config LoggerStruct) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		panic("invalid log level: " + config.Level)
	}

	stdoutCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	fileWriter := &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.FileMaxSizeMB,
		MaxBackups: config.FileMaxBackups,
		MaxAge:     config.FileMaxAgeDays,
		Compress:   true,
	}
	fileCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(fileWriter),
		level,
	)

	combinedCore := zapcore.NewTee(stdoutCore, fileCore)

	opts := []zap.Option{}
	if !config.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}
	if !config.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return zap.New(combinedCore, opts...)
}
