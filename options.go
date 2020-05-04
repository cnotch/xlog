// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import "strings"

// An Option configures a Logger.
type Option interface {
	apply(*Logger)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*Logger)

func (f optionFunc) apply(log *Logger) {
	f(log)
}

// Named adds a new path segment to the logger's name.
// Segments are joined by periods with '.'.
func Named(s string) Option {
	return optionFunc(func(log *Logger) {
		if s == "" {
			return
		}

		if log.name == "" {
			log.name = s
		} else {
			log.name = strings.Join([]string{log.name, s}, ".")
		}
	})
}

// Fields adds preset fields to the Logger.
func Fields(fs ...Field) Option {
	return optionFunc(func(log *Logger) {
		if len(fs) == 0 {
			return
		}
		log.ctx = append(log.ctx, fs...)
	})
}

// AddCaller configures the Logger to annotate each message with the filename
// and line number of caller.
func AddCaller() Option {
	return optionFunc(func(log *Logger) {
		log.addCaller = true
	})
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option).
//
// When building wrappers around the Logger, use this option to set the wrapping depth.
func AddCallerSkip(skip int) Option {
	return optionFunc(func(log *Logger) {
		log.callerSkip += skip
	})
}
