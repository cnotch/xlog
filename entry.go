// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var callerCache sync.Map // map[string]string

// Entry represents a log entry.
type Entry struct {
	Level      Level
	Time       time.Time
	Caller     EntryCaller
	Message    string
	Fields     []Field
	LoggerName string
	Ctx        []Field
}

// EntryCaller represents the caller of a logging function.
type EntryCaller struct {
	Defined bool
	PC      uintptr
	File    string
	Line    int
}

// NewEntryCaller makes an EntryCaller from the return signature of
// runtime.Caller.
func NewEntryCaller(pc uintptr, file string, line int, ok bool) EntryCaller {
	if !ok {
		return EntryCaller{true, 0, "???", 0}
	}

	if ifile, ok := callerCache.Load(file); ok {
		file = ifile.(string)
	} else {
		shortFile := path.Base(file)
		key := file
		funcName := runtime.FuncForPC(pc).Name()
		if i := strings.LastIndexByte(funcName, '/'); i >= 0 {
			i++
			if j := strings.IndexByte(funcName[i:], '.'); j >= 0 {
				i += j
				file = funcName[:i] + "/" + shortFile
			}
		}
		callerCache.Store(key, file)
	}
	return EntryCaller{
		PC:      pc,
		File:    file,
		Line:    line,
		Defined: true,
	}
}

// O represents an object consisting of fields.
type O []Field

// Field represents a custom fielda of log entry.
type Field struct {
	Key string
	Val interface{}
}

// F .
func F(key string, val interface{}) Field {
	return Field{key, val}
}

// String .
func (f Field) String() string {
	var b Builder
	f.appendTo(&b)
	return b.String()
}

// MarshalJSON implements the Marshaler interface.
func (f Field) MarshalJSON() ([]byte, error) {
	var b Builder
	f.appendTo(&b)
	return b.Bytes(), nil
}

func (f Field) appendTo(b *Builder) {
	// key
	b.AppendQuote(f.Key)
	// KV join
	b.WriteByte(':')
	switch v := f.Val.(type) {
	case Field:
		b.WriteByte('{')
		v.appendTo(b)
		b.WriteByte('}')
	case O: // Object
		b.WriteByte('{')
		v.appendTo(b)
		b.WriteByte('}')
	case []O:
		b.WriteByte('[')
		for i, fs := range v {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('{')
			fs.appendTo(b)
			b.WriteByte('}')
		}
		b.WriteByte(']')
	default:
		// value
		b.AppendJSON(f.Val)
	}
}

// MarshalJSON implements the Marshaler interface.
func (o O) MarshalJSON() ([]byte, error) {
	var b Builder
	b.WriteByte('{')
	o.appendTo(&b)
	b.WriteByte('{')
	return b.Bytes(), nil
}

func (o O) appendTo(b *Builder) {
	for i, f := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		f.appendTo(b)
	}
}
