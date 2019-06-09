package log

import (
	"os"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
)

// Path is the default path for where athena data should be stored. This is the log path on
// a k8s pod
const Path = "/var/log/athena"

// DevPath is the log path if we are local and deving (i.e. osx)
const DevPath = "/tmp"

// GetLogPath returns where custom Athena data should be stored,
// i.e. athena errors, parsed sql errors, etc
func getLogPath() string {
	path := os.Getenv("ATHENA_LOG_PATH")
	if path == "" {
		if runtime.GOOS == "darwin" {
			return DevPath
		}
		return Path
	}
	return path
}

func init() {
	// Define our level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	allPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})

	// Get the files we want to log to
	logPath := getLogPath()

	name := filepath.Join(logPath, "debug.log")
	dbgFile, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	name = filepath.Join(logPath, "error.log")
	errFile, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	topicDebugging := zapcore.AddSync(dbgFile)
	topicErrors := zapcore.AddSync(errFile)

	// Let's also log to stderr
	consoleLog := zapcore.Lock(os.Stderr)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// Create development configs
	dbgEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	errEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// Create our custom loggers
	core := zapcore.NewTee(
		zapcore.NewCore(dbgEncoder, topicDebugging, allPriority),
		zapcore.NewCore(errEncoder, topicErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleLog, allPriority),
	)
	log := zap.New(core)
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
