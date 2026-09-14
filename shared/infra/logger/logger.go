package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	Level             string
	EncodingJSON      bool
	EncodingConsole   bool
	DisableCaller     bool
	DisableStacktrace bool

	FilePath       string
	FileMaxSizeMB  int
	FileMaxBackups int
	FileMaxAgeDays int
}

func New(config LoggerConfig) (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder, err := newEncoder(config.EncodingJSON, config.EncodingConsole, encoderConfig)
	if err != nil {
		return nil, err
	}

	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %q, error: %w", config.Level, err)
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

	return zap.New(combinedCore, opts...), nil
}

func newEncoder(useJSON, useConsole bool, cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	if useJSON && useConsole {
		return nil, fmt.Errorf("invalid log config: EncodingJSON and EncodingConsole both are set to true")
	}
	if useConsole {
		return zapcore.NewConsoleEncoder(cfg), nil
	}
	return zapcore.NewJSONEncoder(cfg), nil
}
