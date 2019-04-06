package log

import (
    "context"

    "github.com/apourchet/hermes"
    "go.uber.org/zap"
)

var (
    logger *zap.SugaredLogger
)

func init() {
    config := zap.NewProductionConfig()
    config.DisableCaller = true
    config.DisableStacktrace = true
    log, err := config.Build()
    if err != nil {
        panic(err)
    }
    logger = log.Sugar()
}

func SetLogger(log *zap.SugaredLogger) {
    logger = log
}

func getLogger() *zap.SugaredLogger {
    return logger
}

func Debugf(format string, args ...interface{}) {
    getLogger().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
    getLogger().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
    getLogger().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
    getLogger().Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
    getLogger().Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
    getLogger().Fatalf(format, args...)
}

func Debug(args ...interface{}) {
    getLogger().Debug(args...)
}

func Info(args ...interface{}) {
    getLogger().Info(args...)
}

func Warn(args ...interface{}) {
    getLogger().Warn(args...)
}

func Error(args ...interface{}) {
    getLogger().Error(args...)
}

func Panic(args ...interface{}) {
    getLogger().Panic(args...)
}

func Fatal(args ...interface{}) {
    getLogger().Fatal(args...)
}

func With(args ...interface{}) *zap.SugaredLogger {
    return getLogger().With(args...)
}

func WithRID(ctx context.Context) *zap.SugaredLogger {
    return getLogger().With("request_id", hermes.GetRequestID(ctx))
}

func InternalError(ctx context.Context, err error) {
    WithRID(ctx).With("error_message", err.Error()).
        With("error_type", "internal").
        Errorf("Internal Server Error")
}
