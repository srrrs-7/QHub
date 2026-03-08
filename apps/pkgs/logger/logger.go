package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

// Init - JSON形式のロガーを初期化
func Init() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)
}

// Debug - デバッグログ
func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

// Info - 情報ログ
func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

// Warn - 警告ログ
func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

// Error - エラーログ
func Error(msg string, args ...any) {
	log.Error(msg, args...)
}
