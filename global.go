// Copyright (c) 2018,TianJin Tomatox  Technology Ltd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"os"
	"sync/atomic"
)

var (
	globalL atomic.Value // holds global logger
)

func init() {
	globalL.Store(New(
		NewCore(NewConsoleEncoder(LstdFlags), Lock(os.Stderr), InfoLevel),
		// NewCore(NewConsoleEncoder(LstdFlags), Lock(ioutil.Discard), InfoLevel),
	))
}

// L returns the global Logger, which can be reconfigured with ReplaceGlobals.
// It's safe for concurrent use.
func L() *Logger {
	return globalL.Load().(*Logger)
}

// ReplaceGlobal replaces the global Logger and SugaredLogger, and returns a
// function to restore the original values. It's safe for concurrent use.
func ReplaceGlobal(logger *Logger) func() {
	prev := L()
	if prev == logger {
		return func() {}
	}

	globalL.Store(logger)
	return func() { ReplaceGlobal(prev) }
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...Field) {
	L().log(2, DebugLevel, msg, nil, fields)
}

// Debugf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at DebugLevel.
// The scene is as follows:
//  1. args == nil, directly use template as message
//  2. template = "", uses fmt.Sprint to construct message
//  3. otherwise, uses fmt.Sprintf to construct message
func Debugf(template string, args ...interface{}) {
	L().log(2, DebugLevel, template, args, nil)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...Field) {
	L().log(2, InfoLevel, msg, nil, fields)
}

// Infof uses template or fmt.Sprint or fmt.Sprintf to log a templated message at InfoLevel.
func Infof(template string, args ...interface{}) {
	L().log(2, InfoLevel, template, args, nil)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...Field) {
	L().log(2, WarnLevel, msg, nil, fields)
}

// Warnf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at WarnLevel.
func Warnf(template string, args ...interface{}) {
	L().log(2, WarnLevel, template, args, nil)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...Field) {
	L().log(2, ErrorLevel, msg, nil, fields)
}

// Errorf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at ErrorLevel.
func Errorf(template string, args ...interface{}) {
	L().log(2, ErrorLevel, template, args, nil)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...Field) {
	L().log(2, PanicLevel, msg, nil, fields)
}

// Panicf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at PanicLevel.
// The logger then panics, even if logging at PanicLevel is disabled.
func Panicf(template string, args ...interface{}) {
	L().log(2, PanicLevel, template, args, nil)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...Field) {
	L().log(2, FatalLevel, msg, nil, fields)
}

// Fatalf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at FatalLevel.
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatalf(template string, args ...interface{}) {
	L().log(2, FatalLevel, template, args, nil)
}
