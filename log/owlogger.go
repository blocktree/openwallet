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

type OWLogger struct {
	Std *logger
}

//NewOWLogger 初始化一个日志工具，以[prefix]前缀
func NewOWLogger(prefix string) *OWLogger {
	l := OWLogger{
		Std: newLogger(prefix),
	}
	return &l
}

// SetPrefix 设置前缀
func (logger *OWLogger)SetPrefix(prefix string) {
	logger.Std.SetPrefix(prefix)
}

// SetLevel 设置打印级别
func (logger *OWLogger)SetLevel(l int) {
	logger.Std.SetLevel(l)
}

// SetLogFuncCall set the CallDepth, default is 3
func (logger *OWLogger)SetLogFuncCall(b bool) {
	logger.Std.EnableFuncCallDepth(b)
	logger.Std.SetLogFuncCallDepth(3)
}

// SetLogger sets a new logger.
func (logger *OWLogger)SetLogger(adaptername string, config string) error {
	return logger.Std.SetLogger(adaptername, config)
}

// Emergency logs a message at emergency level.
func (logger *OWLogger)Emergency(v ...interface{}) {
	logger.Std.Emergency(generateFmtStr(len(v)), v...)
}

// Alert logs a message at alert level.
func (logger *OWLogger)Alert(v ...interface{}) {
	logger.Std.Alert(generateFmtStr(len(v)), v...)
}

// Critical logs a message at critical level.
func (logger *OWLogger)Critical(v ...interface{}) {
	SetLogFuncCall(true)
	logger.Std.Critical(generateFmtStr(len(v)), v...)
}

// format & Error logs a message at error level.
func (logger *OWLogger)Errorf(format string, v ...interface{}) {
	SetLogFuncCall(true)
	logger.Std.Error(format, v...)
}

// Error logs a message at error level.
func (logger *OWLogger)Error(v ...interface{}) {
	SetLogFuncCall(true)
	logger.Std.Error(generateFmtStr(len(v)), v...)
}

// format & Warning logs a message at warning level.
func (logger *OWLogger)Warningf(format string, v ...interface{}) {
	logger.Std.Warning(format, v...)
}

// Warning logs a message at warning level.
func (logger *OWLogger)Warning(v ...interface{}) {
	logger.Std.Warning(generateFmtStr(len(v)), v...)
}

// Warn compatibility alias for Warning()
func (logger *OWLogger)Warn(v ...interface{}) {
	logger.Std.Warn(generateFmtStr(len(v)), v...)
}

// Notice logs a message at notice level.
func (logger *OWLogger)Notice(v ...interface{}) {
	logger.Std.Notice(generateFmtStr(len(v)), v...)
}

// Informational logs a message at info level.
func (logger *OWLogger)Informational(v ...interface{}) {
	logger.Std.Informational(generateFmtStr(len(v)), v...)
}

func (logger *OWLogger)Infof(format string, v ...interface{}) {
	logger.Std.Info(format, v...)
}

// Info compatibility alias for Warning()
func (logger *OWLogger)Info(v ...interface{}) {
	logger.Std.Info(generateFmtStr(len(v)), v...)
}

// Format & debug logs a message at debug level.
func (logger *OWLogger)Debugf(format string, v ...interface{}) {
	SetLogFuncCall(true)
	logger.Std.Debug(format, v...)
}

// Debug logs a message at debug level.
func (logger *OWLogger)Debug(v ...interface{}) {
	SetLogFuncCall(true)
	logger.Std.Debug(generateFmtStr(len(v)), v...)
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func (logger *OWLogger)Trace(v ...interface{}) {
	logger.Std.Trace(generateFmtStr(len(v)), v...)
}

func (logger *OWLogger)generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}

