package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

// InitLogger initializes the global logger
func InitLogger() {
	config := zap.NewProductionConfig()

	// Set more readable time format
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create logger
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	// Use SugaredLogger for easier key-value logging
	Log = logger.Sugar()
}

// InitLoggerDev initializes logger in development mode (more readable output)
func InitLoggerDev() {
	config := zap.NewDevelopmentConfig()

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	Log = logger.Sugar()
}

// Sync flushes buffered logs
func Sync() {
	if Log != nil {
		Log.Sync()
	}
}
