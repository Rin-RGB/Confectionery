package core_logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"SPOproject/internal/core/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	file *os.File
}

type contextKey struct{}

func NewLogger(config config.LoggerConfig) (*Logger, error) {
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("parse logger level: %w", err)
	}
	if err = os.MkdirAll(config.Folder, 0o750); err != nil {
		return nil, fmt.Errorf("create logger folder: %w", err)
	}

	fileName := time.Now().UTC().Format("2006-01-02T15-04-05.000000") + ".log"
	filePath := filepath.Join(config.Folder, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", filePath, err)
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	atomicLevel := zap.NewAtomicLevelAt(level)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atomicLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(file), atomicLevel),
	)

	return &Logger{
		Logger: zap.New(core, zap.AddCaller()),
		file:   file,
	}, nil
}

func (logger *Logger) Close() error {
	if err := logger.Sync(); err != nil && !isInvalidSyncError(err) {
		return fmt.Errorf("sync logger: %w", err)
	}
	if err := logger.file.Close(); err != nil {
		return fmt.Errorf("close log file: %w", err)
	}
	return nil
}

func (logger *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: logger.Logger.With(fields...),
		file:   logger.file,
	}
}

func CtxWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func FromContext(ctx context.Context) *Logger {
	contextLogger, ok := ctx.Value(contextKey{}).(*Logger)
	if !ok {
		panic("logger is missing from context")
	}
	return contextLogger
}

func isInvalidSyncError(err error) bool {
	return err.Error() == "sync /dev/stdout: invalid argument" ||
		err.Error() == "sync /dev/stderr: invalid argument"
}
