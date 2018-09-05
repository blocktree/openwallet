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

/********** Modified from beego **********/

// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package openwLogger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/blocktree/OpenWallet/logger/colors"
)

var errInvalidLogLevel = errors.New("logger: invalid log level")

const (
	levelDebug = iota
	levelError
	levelFatal
	levelCritical
	levelSuccess
	levelWarn
	levelInfo
	levelHint
)

var (
	sequenceNo uint64
	instance   *OpenwLogger
	once       sync.Once
)
var debugMode = os.Getenv("DEBUG_ENABLED") == "1"

var logLevel = levelInfo

// OpenwLogger logs logging records to the specified io.Writer
// Deprecated: use github.com/blocktree/OpenWallet/log/
type OpenwLogger struct {
	mu     sync.Mutex
	output io.Writer
}

// LogRecord represents a log record and contains the timestamp when the record
// was created, an increasing id, level and the actual formatted log line.
type LogRecord struct {
	ID       string
	Level    string
	Message  string
	Filename string
	LineNo   int
}

// Deprecated: use github.com/blocktree/OpenWallet/log/ => Log
var Log = GetOpenwLogger(os.Stdout)

var (
	logRecordTemplate      *template.Template
	debugLogRecordTemplate *template.Template
)

// GetOpenwLogger initializes the logger instance with a NewColorWriter output
// and returns a singleton
func GetOpenwLogger(w io.Writer) *OpenwLogger {
	once.Do(func() {
		var (
			err             error
			simpleLogFormat = `{{Now "2006/01/02 15:04:05"}} {{.Level}} ▶ {{.ID}} {{.Message}}{{EndLine}}`
			debugLogFormat  = `{{Now "2006/01/02 15:04:05"}} {{.Level}} ▶ {{.ID}} {{.Filename}}:{{.LineNo}} {{.Message}}{{EndLine}}`
		)

		// Initialize and parse logging templates
		funcs := template.FuncMap{
			"Now":     Now,
			"EndLine": EndLine,
		}
		logRecordTemplate, err = template.New("simpleLogFormat").Funcs(funcs).Parse(simpleLogFormat)
		if err != nil {
			panic(err)
		}
		debugLogRecordTemplate, err = template.New("debugLogFormat").Funcs(funcs).Parse(debugLogFormat)
		if err != nil {
			panic(err)
		}

		instance = &OpenwLogger{output: colors.NewColorWriter(w)}
	})
	return instance
}

// SetOutput sets the logger output destination
func (l *OpenwLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = colors.NewColorWriter(w)
}

// Now returns the current local time in the specified layout
func Now(layout string) string {
	return time.Now().Format(layout)
}

// EndLine returns the a newline escape character
func EndLine() string {
	return "\n"
}

func (l *OpenwLogger) getLevelTag(level int) string {
	switch level {
	case levelFatal:
		return "FATAL   "
	case levelSuccess:
		return "SUCCESS "
	case levelHint:
		return "HINT    "
	case levelDebug:
		return "DEBUG   "
	case levelInfo:
		return "INFO    "
	case levelWarn:
		return "WARN    "
	case levelError:
		return "ERROR   "
	case levelCritical:
		return "CRITICAL"
	default:
		panic(errInvalidLogLevel)
	}
}

func (l *OpenwLogger) getColorLevel(level int) string {
	switch level {
	case levelCritical:
		return colors.RedBold(l.getLevelTag(level))
	case levelFatal:
		return colors.RedBold(l.getLevelTag(level))
	case levelInfo:
		return colors.BlueBold(l.getLevelTag(level))
	case levelHint:
		return colors.CyanBold(l.getLevelTag(level))
	case levelDebug:
		return colors.YellowBold(l.getLevelTag(level))
	case levelError:
		return colors.RedBold(l.getLevelTag(level))
	case levelWarn:
		return colors.YellowBold(l.getLevelTag(level))
	case levelSuccess:
		return colors.GreenBold(l.getLevelTag(level))
	default:
		panic(errInvalidLogLevel)
	}
}

// mustLog logs the message according to the specified level and arguments.
// It panics in case of an error.
func (l *OpenwLogger) mustLog(level int, message string, args ...interface{}) {
	if level > logLevel {
		return
	}
	// Acquire the lock
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create the logging record and pass into the output
	record := LogRecord{
		ID:      fmt.Sprintf("%04d", atomic.AddUint64(&sequenceNo, 1)),
		Level:   l.getColorLevel(level),
		Message: fmt.Sprintf(message, args...),
	}

	err := logRecordTemplate.Execute(l.output, record)
	if err != nil {
		panic(err)
	}
}

// mustLogDebug logs a debug message only if debug mode
// is enabled. i.e. DEBUG_ENABLED="1"
func (l *OpenwLogger) mustLogDebug(message string, file string, line int, args ...interface{}) {
	if !debugMode {
		return
	}

	// Change the output to Stderr
	l.SetOutput(os.Stderr)

	// Create the log record
	record := LogRecord{
		ID:       fmt.Sprintf("%04d", atomic.AddUint64(&sequenceNo, 1)),
		Level:    l.getColorLevel(levelDebug),
		Message:  fmt.Sprintf(message, args...),
		LineNo:   line,
		Filename: filepath.Base(file),
	}
	err := debugLogRecordTemplate.Execute(l.output, record)
	if err != nil {
		panic(err)
	}
}

// Debug outputs a debug log message
func (l *OpenwLogger) Debug(message string, file string, line int) {
	l.mustLogDebug(message, file, line)
}

// Debugf outputs a formatted debug log message
func (l *OpenwLogger) Debugf(message string, vars ...interface{}) {
	pc, file, lineno, ok := runtime.Caller(1)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		tokens := strings.Split(funcName, ".")
		funcName = tokens[len(tokens)-1]
		tokens = strings.Split(file, "/")
		file = tokens[len(tokens)-1]
		l.mustLogDebug("["+funcName+"] "+message, file, lineno, vars...)
	}

}

// Info outputs an information log message
func (l *OpenwLogger) Info(message string) {
	l.mustLog(levelInfo, message)
}

// Infof outputs a formatted information log message
func (l *OpenwLogger) Infof(message string, vars ...interface{}) {
	l.mustLog(levelInfo, message, vars...)
}

// Warn outputs a warning log message
func (l *OpenwLogger) Warn(message string) {
	l.mustLog(levelWarn, message)
}

// Warnf outputs a formatted warning log message
func (l *OpenwLogger) Warnf(message string, vars ...interface{}) {
	l.mustLog(levelWarn, message, vars...)
}

// Error outputs an error log message
func (l *OpenwLogger) Error(message string) {
	l.mustLog(levelError, message)
}

// Errorf outputs a formatted error log message
func (l *OpenwLogger) Errorf(message string, vars ...interface{}) {
	l.mustLog(levelError, message, vars...)
}

// Fatal outputs a fatal log message and exists
func (l *OpenwLogger) Fatal(message string) {
	l.mustLog(levelFatal, message)
	os.Exit(255)
}

// Fatalf outputs a formatted log message and exists
func (l *OpenwLogger) Fatalf(message string, vars ...interface{}) {
	l.mustLog(levelFatal, message, vars...)
	os.Exit(255)
}

// Success outputs a success log message
func (l *OpenwLogger) Success(message string) {
	l.mustLog(levelSuccess, message)
}

// Successf outputs a formatted success log message
func (l *OpenwLogger) Successf(message string, vars ...interface{}) {
	l.mustLog(levelSuccess, message, vars...)
}

// Hint outputs a hint log message
func (l *OpenwLogger) Hint(message string) {
	l.mustLog(levelHint, message)
}

// Hintf outputs a formatted hint log message
func (l *OpenwLogger) Hintf(message string, vars ...interface{}) {
	l.mustLog(levelHint, message, vars...)
}

// Critical outputs a critical log message
func (l *OpenwLogger) Critical(message string) {
	l.mustLog(levelCritical, message)
}

// Criticalf outputs a formatted critical log message
func (l *OpenwLogger) Criticalf(message string, vars ...interface{}) {
	l.mustLog(levelCritical, message, vars...)
}
