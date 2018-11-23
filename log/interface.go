package log

const module = "module"
const (
	ZapLogger     = "zap"
	Log15Logger   = "log15"
	DefaultLogger = Log15Logger
)

var baseLogger Logger

type Logger interface {
	NewWith(keysAndValues ...interface{}) Logger
	newModuleLogger(mi ModuleID) Logger
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

func init() {
	root.SetHandler(DiscardHandler())
	SetBaseLogger()
}

func SetBaseLogger() {
	switch DefaultLogger {
	case ZapLogger:
		baseLogger = genBaseLoggerZap()
	case Log15Logger:
		baseLogger = root
	default:
		baseLogger = genBaseLoggerZap()
	}
}

func NewModuleLogger(mi ModuleID) Logger {
	newLogger := baseLogger.newModuleLogger(mi)
	return newLogger
}
