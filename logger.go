// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// A Logger provides fast, leveled, structured logging.
// All methods are safe for concurrent use.
type Logger struct {
	core       Core
	addCaller  bool
	callerSkip int
	name       string
	ctx        []Field
}

// New constructs a new Logger from the provided Core and Options.
// If the passed Core is nil, it falls back to using a no-op implementation.
func New(core Core, options ...Option) *Logger {
	if core == nil {
		core = NewNopCore()
	}

	log := &Logger{
		core: core,
	}

	for _, opt := range options {
		opt.apply(log)
	}
	return log
}

// With clones the current Logger, applies the supplied Options, and
// returns the resulting Logger. It's safe to use concurrently.
func (l *Logger) With(opts ...Option) *Logger {
	c := l.clone()
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

// LevelEnabled 日志对象指定的级别是否启用
func (l *Logger) LevelEnabled(lvl Level) bool {
	if lvl < DebugLevel || lvl > FatalLevel {
		return false
	}
	return l.core.Enabled(lvl)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(2, DebugLevel, msg, nil, fields)
}

// Debugf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at DebugLevel.
// The scene is as follows:
//  1. args == nil, directly use template as message
//  2. template = "", uses fmt.Sprint to construct message
//  3. otherwise, uses fmt.Sprintf to construct message
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.log(2, DebugLevel, template, args, nil)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(2, InfoLevel, msg, nil, fields)
}

// Infof uses template or fmt.Sprint or fmt.Sprintf to log a templated message at InfoLevel.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.log(2, InfoLevel, template, args, nil)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(2, WarnLevel, msg, nil, fields)
}

// Warnf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at WarnLevel.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.log(2, WarnLevel, template, args, nil)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(2, ErrorLevel, msg, nil, fields)
}

// Errorf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at ErrorLevel.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.log(2, ErrorLevel, template, args, nil)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (l *Logger) Panic(msg string, fields ...Field) {
	l.log(2, PanicLevel, msg, nil, fields)
}

// Panicf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at PanicLevel.
// The logger then panics, even if logging at PanicLevel is disabled.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.log(2, PanicLevel, template, args, nil)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(2, FatalLevel, msg, nil, fields)
}

// Fatalf uses template or fmt.Sprint or fmt.Sprintf to log a templated message at FatalLevel.
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.log(2, FatalLevel, template, args, nil)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (l *Logger) Sync() error {
	return l.core.Sync()
}

// Core returns the Logger's underlying Core.
func (l *Logger) Core() Core {
	return l.core
}

// all logical of log op.
func (l *Logger) log(calloffset int, lvl Level, template string, fmtArgs []interface{}, fields []Field) {
	if !l.core.Enabled(lvl) {
		switch lvl {
		case PanicLevel:
			panic(messagef(template, fmtArgs...))
		case FatalLevel:
			os.Exit(1)
		}
		return
	}

	e := Entry{
		Level:      lvl,
		Time:       time.Now(),
		Message:    messagef(template, fmtArgs...),
		Fields:     fields,
		LoggerName: l.name,
		Ctx:        l.ctx,
	}

	if l.addCaller {
		e.Caller = NewEntryCaller(runtime.Caller(l.callerSkip + calloffset))
	}

	if err := l.core.Write(e); err != nil {
		// TODO: handle internal log errors
	}

	// PanicLevel and FatalLevel require additional operations
	switch lvl {
	case PanicLevel:
		panic(e.Message)
	case FatalLevel:
		os.Exit(1)
	}
}

func (l *Logger) clone() *Logger {
	c := *l
	c.ctx = nil
	// avoid the subsequent addition of preset fields to interfere with l
	c.ctx = append(c.ctx, l.ctx...)
	return &c
}

func messagef(template string, args ...interface{}) string {
	// Format with Sprint, Sprintf, or neither.
	if template == "" && len(args) > 0 {
		return fmt.Sprint(args...)
	} else if template != "" && len(args) > 0 {
		return fmt.Sprintf(template, args...)
	}
	return template
}
