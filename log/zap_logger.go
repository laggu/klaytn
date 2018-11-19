package log

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

var (
	defaultLogOutputPath = ""
	defaultLogOutputFile = "klaytn-log"
	defaultLogEncodingType = "json"
	defaultLogLevel = zapcore.InfoLevel
	defaultMessageKey = "msg"
	defaultLoggerName = "defaultlogger"
	moduleConfigMap = make(map[string]*zap.Config)
)

func genDefaultEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		//CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		//EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func getDefaultConfig() *zap.Config {
	return moduleConfigMap[defaultLoggerName]
}

func genDefaultConfig() *zap.Config {
	encoderConfig := genDefaultEncoderConfig()
	return &zap.Config{
		Encoding:    defaultLogEncodingType,
		Level:       zap.NewAtomicLevelAt(defaultLogLevel),
		OutputPaths: []string{defaultLogOutputPath},
		Development: false,
		EncoderConfig: encoderConfig,
	}
}

func genBaseLoggerZap() Logger {
	ex, err := os.Executable()
	if err != nil {
		// TODO-GX Error should be handled.
	}
	// TODO-GX Output path should be set properly.
	defaultLogOutputPath = path.Join(filepath.Dir(ex), defaultLogOutputFile)

	// defaultLogger is bound to package level log function.
	// This is to have consistent logging behavior between transition period.
	defaultLoggerCfg := genDefaultConfig()
	logger, err := defaultLoggerCfg.Build()
	if err != nil {
		// TODO-GX Error should be handled.
	}

	moduleConfigMap[defaultLoggerName] = defaultLoggerCfg
	return &zapLogger{logger.Sugar()}
}

type Logger interface {
	New(keysAndValues ...interface{}) Logger
	Trace(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Crit(msg string, keysAndValues ...interface{})

	// GetHandler gets the handler associated with the logger.
	GetHandler() Handler
	// SetHandler updates the logger to write records to the specified handler.
	SetHandler(h Handler)
}

type zapLogger struct {
	sl *zap.SugaredLogger
}

func (zl *zapLogger) New(keysAndValues ...interface{}) Logger {
	return &zapLogger{zl.sl.With(keysAndValues...)}
}

func (zl *zapLogger) Trace(msg string, keysAndValues ...interface{}) {
	zl.sl.Debugw(msg, keysAndValues...)
}

func (zl *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	zl.sl.Debugw(msg, keysAndValues...)
}

func (zl *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	zl.sl.Infow(msg, keysAndValues...)
}

func (zl *zapLogger) Warn(msg string, keysAndValues ...interface{}) {
	zl.sl.Warnw(msg, keysAndValues...)
}

func (zl *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	zl.sl.Errorw(msg, keysAndValues...)
}

func (zl *zapLogger) Crit(msg string, keysAndValues ...interface{}) {
	zl.sl.Fatalw(msg, keysAndValues...)
}

// GetHandler and SetHandler do nothing but exist to make consistency in Logger interface.
func (zl *zapLogger) GetHandler() Handler {
	return nil
}

func (zl *zapLogger) SetHandler(h Handler) {
}

func NewLogger(moduleName string) Logger {
	moduleName = strings.ToLower(moduleName)

	if moduleConfigMap[moduleName] != nil {
		Crit("Duplicated log moduleName found!", "moduleName", moduleName)
	}

	zapCfg := genDefaultConfig()
	zapCfg.InitialFields["module"] = moduleName
	logger, err := zapCfg.Build()
	if err != nil {
		// TODO-GX Error should be handled.
	}

	moduleConfigMap[moduleName] = zapCfg
	return &zapLogger{logger.Sugar()}
}

func ChangeLogLevel(moduleName string, lvl Lvl) error {
	moduleName = strings.ToLower(moduleName)
	cfg := moduleConfigMap[moduleName]
	if cfg == nil {
		return errors.New("entered module name does not match with any existing log module")
	}

	cfg.Level.SetLevel(lvlToZapLevel(lvl))
	return nil
}

func ChangeGlobalLogLevel(lvl Lvl) {
	for _, cfg := range moduleConfigMap {
		cfg.Level.SetLevel(lvlToZapLevel(lvl))
	}
}

func lvlToZapLevel(lvl Lvl) zapcore.Level {
	switch lvl {
	case LvlCrit:
		return zapcore.FatalLevel
	case LvlError:
		return zapcore.ErrorLevel
	case LvlWarn:
		return zapcore.WarnLevel
	case LvlInfo:
		return zapcore.InfoLevel
	case LvlDebug:
		return zapcore.DebugLevel
	case LvlTrace:
		return zapcore.DebugLevel
	default:
		Error("Unexpected log level entered. Use InfoLevel instead.", "entered level", lvl)
		return zapcore.InfoLevel
	}
}