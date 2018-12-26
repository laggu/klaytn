// Copyright 2018 The go-klaytn Authors
//
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

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
