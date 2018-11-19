package log

const (
	ZapLogger     = "ZapLogger"
	Log15Logger   = "Log15Logger"
	DefaultLogger = Log15Logger
)

var BaseLogType = DefaultLogger
var baseLogger Logger

func init() {
	root.SetHandler(DiscardHandler())
	SetBaseLogger(BaseLogType)
}

func SetBaseLogger(logType string) {
	switch logType {
	case ZapLogger:
		baseLogger = genBaseLoggerZap()
	case Log15Logger:
		baseLogger = root
	default:
		baseLogger = genBaseLoggerZap()
	}
}

// New returns a new logger with the given context.
func New(keyAndValues ...interface{}) Logger {
	return baseLogger.New(keyAndValues...)
}

func Trace(msg string, keyAndValues ...interface{}) {
	baseLogger.Trace(msg, keyAndValues...)
}

func Debug(msg string, keyAndValues ...interface{}) {
	baseLogger.Debug(msg, keyAndValues...)
}

func Info(msg string, keyAndValues ...interface{}) {
	baseLogger.Info(msg, keyAndValues...)
}

func Warn(msg string, keyAndValues ...interface{}) {
	baseLogger.Warn(msg, keyAndValues...)
}

func Error(msg string, keyAndValues ...interface{}) {
	baseLogger.Error(msg, keyAndValues...)
}

func Crit(msg string, keyAndValues ...interface{}) {
	baseLogger.Crit(msg, keyAndValues...)
}
