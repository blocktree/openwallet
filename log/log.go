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

package log

import (
	"strings"
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
var Std = newLogger("")

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	Std.SetLevel(l)
}

// SetLogFuncCall set the CallDepth, default is 3
func SetLogFuncCall(b bool) {
	Std.EnableFuncCallDepth(b)
	Std.SetLogFuncCallDepth(4)
}

// SetLogger sets a new logger.
func SetLogger(adaptername string, config string) error {
	return Std.SetLogger(adaptername, config)
}

// Emergency logs a message at emergency level.
func Emergency(v ...interface{}) {
	Std.Emergency(generateFmtStr(len(v)), v...)
}

// Alert logs a message at alert level.
func Alert(v ...interface{}) {
	Std.Alert(generateFmtStr(len(v)), v...)
}

// Critical logs a message at critical level.
func Critical(v ...interface{}) {
	SetLogFuncCall(true)
	Std.Critical(generateFmtStr(len(v)), v...)
}

// format & Error logs a message at error level.
func Errorf(format string, v ...interface{}) {
	SetLogFuncCall(true)
	Std.Error(format, v...)
}

// Error logs a message at error level.
func Error(v ...interface{}) {
	SetLogFuncCall(true)
	Std.Error(generateFmtStr(len(v)), v...)
}

// format & Warning logs a message at warning level.
func Warningf(format string, v ...interface{}) {
	Std.Warning(format, v...)
}

// Warning logs a message at warning level.
func Warning(v ...interface{}) {
	Std.Warning(generateFmtStr(len(v)), v...)
}

// Warn compatibility alias for Warning()
func Warn(v ...interface{}) {
	Std.Warn(generateFmtStr(len(v)), v...)
}

// Notice logs a message at notice level.
func Notice(v ...interface{}) {
	Std.Notice(generateFmtStr(len(v)), v...)
}

// Informational logs a message at info level.
func Informational(v ...interface{}) {
	Std.Informational(generateFmtStr(len(v)), v...)
}

func Infof(format string, v ...interface{}) {
	Std.Info(format, v...)
}

// Info compatibility alias for Warning()
func Info(v ...interface{}) {
	Std.Info(generateFmtStr(len(v)), v...)
}

// Format & debug logs a message at debug level.
func Debugf(format string, v ...interface{}) {
	SetLogFuncCall(true)
	Std.Debug(format, v...)
}

// Debug logs a message at debug level.
func Debug(v ...interface{}) {
	SetLogFuncCall(true)
	Std.Debug(generateFmtStr(len(v)), v...)
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(v ...interface{}) {
	Std.Trace(generateFmtStr(len(v)), v...)
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}
