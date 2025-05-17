package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a zap.Logger based on the desired level.
// In "dev" we use Console encoder; in "prod", JSON.
func NewLogger(env, level string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "prod" {
		cfg = zap.NewProductionConfig()
		cfg.Encoding = "json"
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.Encoding = "console"
		// In dev, keep things simple—no stacktraces on warnings
		cfg.DisableStacktrace = true
	}
	// Parse and set level
	lvl := zapcore.InfoLevel
	if err := lvl.Set(level); err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	return cfg.Build()
}
