package log

import (
	"fmt"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/syncmap"
)

// Global variables against which all our logging occurs.
var logger *zap.Logger
var sugar *zap.SugaredLogger
var goRoutineMapping = syncmap.Map{}

type routineVariables struct {
	UUID string
	IP   string
}

func init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "Time",
		LevelKey:       "Level",
		NameKey:        "Name",
		MessageKey:     "Message",
		StacktraceKey:  "Stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	atom := zap.NewAtomicLevelAt(zap.InfoLevel)

	config := zap.Config{
		Level:            atom,
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("log init err: %v", err))
	}
	logger = l
	sugar = logger.Sugar()
}

func AddRoutineVariable(uuid string, ip string) {
	goRoutineMapping.Store(GetCurrentGoroutineID(), routineVariables{UUID: uuid, IP: ip})
}

// Debug outputs a message at debug level.
// This call is a wrapper around [Logger.Debug](https://godoc.org/go.uber.org/zap#Logger.Debug)
func ZapDebug(msg string, fields ...zapcore.Field) {
	fields = getZapFields(fields)
	logger.Debug(msg, fields...)
}

// DebugEnabled returns whether output of messages at the debug level is currently enabled.
func DebugEnabled() bool {
	return logger.Core().Enabled(zap.DebugLevel)
}

// Error outputs a message at error level.
// This call is a wrapper around [logger.Error](https://godoc.org/go.uber.org/zap#logger.Error)
func ZapError(msg string, fields ...zapcore.Field) {
	fields = getZapFields(fields)
	logger.Error(msg, fields...)
}

// ErrorEnabled returns whether output of messages at the error level is currently enabled.
func ErrorEnabled() bool {
	return logger.Core().Enabled(zap.ErrorLevel)
}

// Warn outputs a message at warn level.
// This call is a wrapper around [logger.Warn](https://godoc.org/go.uber.org/zap#logger.Warn)
func ZapWarn(msg string, fields ...zapcore.Field) {
	fields = getZapFields(fields)
	logger.Warn(msg, fields...)
}

// WarnEnabled returns whether output of messages at the warn level is currently enabled.
func WarnEnabled() bool {
	return logger.Core().Enabled(zap.WarnLevel)
}

// Info outputs a message at information level.
// This call is a wrapper around [logger.Info](https://godoc.org/go.uber.org/zap#logger.Info)
func ZapInfo(msg string, fields ...zapcore.Field) {
	fields = getZapFields(fields)
	logger.Info(msg, fields...)
}

// InfoEnabled returns whether output of messages at the info level is currently enabled.
func InfoEnabled() bool {
	return logger.Core().Enabled(zap.InfoLevel)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
// This call is a wrapper around [logger.With](https://godoc.org/go.uber.org/zap#logger.With)
func With(fields ...zapcore.Field) *zap.Logger {
	fields = getZapFields(fields)
	return logger.With(fields...)
}

// Sync flushes any buffered log entries.
// Processes should normally take care to call Sync before exiting.
// This call is a wrapper around [logger.Sync](https://godoc.org/go.uber.org/zap#logger.Sync)
func Sync() error {
	return logger.Sync()
}

func Debug(args ...interface{}) {
	sugar.Debug(getExtraInfo(), args)
}
func Info(args ...interface{}) {
	sugar.Info(getExtraInfo(), args)
}

func Print(args ...interface{}) {
	sugar.Info(getExtraInfo(), args)
}
func Println(args ...interface{}) {
	sugar.Info(getExtraInfo(), args)
}
func Infof(template string, args ...interface{}) {
	sugar.Infof(getExtraInfo()+template, args...)
}
func Printf(template string, args ...interface{}) {
	sugar.Infof(getExtraInfo()+template, args...)
}
func Infow(msg string, keysAndValues ...interface{}) {
	sugar.Infow(getExtraInfo()+msg, keysAndValues...)
}
func Warn(args ...interface{}) {
	sugar.Warn(getExtraInfo(), args)
}
func Error(args ...interface{}) {
	sugar.Error(getExtraInfo(), args)
}
func Panic(args ...interface{}) {
	sugar.Panic(getExtraInfo(), args)
}
func Fatal(args ...interface{}) {
	sugar.Fatal(getExtraInfo(), args)
}
func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(getExtraInfo()+template, args...)
}

func getZapFields(fields []zapcore.Field) (resultFields []zapcore.Field) {
	uuid, ip := getContextInfo()
	fileName, funcName, lineNumber := getFuncDetails()
	if uuid == "" {
		resultFields = append(fields, zap.String("funcName", funcName), zap.Int("lineNumber", lineNumber), zap.String("fileName", fileName))
	} else {
		resultFields = append(fields, zap.String("funcName", funcName), zap.Int("lineNumber", lineNumber), zap.String("RequestID", uuid), zap.String("IP", ip), zap.String("fileName", fileName))
	}
	return
}

func getExtraInfo() (info string) {
	uuid, ip := getContextInfo()
	fileName, funcName, lineNumber := getFuncDetails()
	if uuid == "" {
		info = fmt.Sprintf("%s:%v @%s ", fileName, lineNumber, funcName)
	} else {
		info = fmt.Sprintf("RequestID:%s IP:%s => %s:%v @%s ", uuid, ip, fileName, lineNumber, funcName)
	}
	return
}

func getContextInfo() (uuid string, ip string) {
	if value, ok := goRoutineMapping.Load(GetCurrentGoroutineID()); ok {
		if mapping, ok := value.(routineVariables); ok {
			uuid, ip = mapping.UUID, mapping.IP
		}
	}
	return
}

func getFuncDetails() (fileName string, funcName string, lineNumber int) {
	pt, file, line, ok := runtime.Caller(3)
	details := runtime.FuncForPC(pt)
	// a := details.Entry()
	// fmt.Println(a)
	if ok && details != nil {
		return filepath.Base(file), filepath.Base(details.Name()), line
	}
	return "", "", 0
}
