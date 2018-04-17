/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwallet

import (
	"strings"

	"github.com/astaxie/beego/logs"
)

// Log levels to control the logging output.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// BeeLogger references the used application logger.
var OPLogger = logs.GetBeeLogger()

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	logs.SetLevel(l)
}

// SetLogFuncCall set the CallDepth, default is 3
func SetLogFuncCall(b bool) {
	logs.SetLogFuncCall(b)
}

// SetLogger sets a new logger.
func SetLogger(adaptername string, config string) error {
	return logs.SetLogger(adaptername, config)
}

// Emergency logs a message at emergency level.
func Emergency(v ...interface{}) {
	logs.Emergency(generateFmtStr(len(v)), v...)
}

// Alert logs a message at alert level.
func Alert(v ...interface{}) {
	logs.Alert(generateFmtStr(len(v)), v...)
}

// Critical logs a message at critical level.
func Critical(v ...interface{}) {
	logs.Critical(generateFmtStr(len(v)), v...)
}

// Error logs a message at error level.
func Error(v ...interface{}) {
	logs.Error(generateFmtStr(len(v)), v...)
}

// Warning logs a message at warning level.
func Warning(v ...interface{}) {
	logs.Warning(generateFmtStr(len(v)), v...)
}

// Warn compatibility alias for Warning()
func Warn(v ...interface{}) {
	logs.Warn(generateFmtStr(len(v)), v...)
}

// Notice logs a message at notice level.
func Notice(v ...interface{}) {
	logs.Notice(generateFmtStr(len(v)), v...)
}

// Informational logs a message at info level.
func Informational(v ...interface{}) {
	logs.Informational(generateFmtStr(len(v)), v...)
}

// Info compatibility alias for Warning()
func Info(v ...interface{}) {
	logs.Info(generateFmtStr(len(v)), v...)
}

// Debug logs a message at debug level.
func Debug(v ...interface{}) {
	logs.Debug(generateFmtStr(len(v)), v...)
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(v ...interface{}) {
	logs.Trace(generateFmtStr(len(v)), v...)
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}
