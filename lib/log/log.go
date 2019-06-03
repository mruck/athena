package log

import (
	"go.uber.org/zap"
)

var (
	logger *zap.SugaredLogger
)

func init() {
	// If I want json logs use a production config
	//config := zap.NewProductionConfig()
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stderr", "/var/log/athena/athena.log"}
	config.DisableCaller = true
	config.DisableStacktrace = false
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
