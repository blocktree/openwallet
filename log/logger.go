/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package log

import (
	"fmt"
	"github.com/astaxie/beego/logs"
)

type logger struct {
	*logs.BeeLogger
	prefix string
}

func newLogger(prefix string) *logger {
	l := logger{}
	if len(prefix) > 0 {
		l.prefix = fmt.Sprintf("[%s] ", prefix)
	}
	l.BeeLogger = logs.NewLogger()
	return &l
}

// SetPrefix
func (bl *logger)SetPrefix(prefix string) {
	if len(prefix) > 0 {
		bl.prefix = fmt.Sprintf("[%s] ", prefix)
	}
}

// Emergency Log EMERGENCY level message.
func (bl *logger) Emergency(format string, v ...interface{}) {
	bl.BeeLogger.Emergency(bl.prefix + format, v...)
}

// Alert Log ALERT level message.
func (bl *logger) Alert(format string, v ...interface{}) {
	bl.BeeLogger.Alert(bl.prefix + format, v...)
}

// Critical Log CRITICAL level message.
func (bl *logger) Critical(format string, v ...interface{}) {
	bl.BeeLogger.Critical(bl.prefix + format, v...)
}

// Error Log ERROR level message.
func (bl *logger) Error(format string, v ...interface{}) {
	bl.BeeLogger.Error(bl.prefix + format, v...)
}

// Warning Log WARNING level message.
func (bl *logger) Warning(format string, v ...interface{}) {
	bl.BeeLogger.Warning(bl.prefix + format, v...)
}

// Notice Log NOTICE level message.
func (bl *logger) Notice(format string, v ...interface{}) {
	bl.BeeLogger.Notice(bl.prefix + format, v...)
}

// Informational Log INFORMATIONAL level message.
func (bl *logger) Informational(format string, v ...interface{}) {
	bl.BeeLogger.Informational(bl.prefix + format, v...)
}

// Debug Log DEBUG level message.
func (bl *logger) Debug(format string, v ...interface{}) {
	bl.BeeLogger.Debug(bl.prefix + format, v...)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (bl *logger) Warn(format string, v ...interface{}) {
	bl.BeeLogger.Warn(bl.prefix + format, v...)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (bl *logger) Info(format string, v ...interface{}) {
	bl.BeeLogger.Info(bl.prefix + format, v...)
}

// Trace Log TRACE level message.
// compatibility alias for Debug()
func (bl *logger) Trace(format string, v ...interface{}) {
	bl.BeeLogger.Trace(bl.prefix + format, v...)
}




