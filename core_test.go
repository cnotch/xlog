// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"bytes"
	"testing"
	"time"
)

func TestCore_Write_console(t *testing.T) {
	cases := []struct {
		e    Entry
		want string
	}{
		{
			Entry{
				Level:      InfoLevel,
				Time:       time.Date(2019, 1, 18, 12, 0, 35, 9876, time.UTC),
				Caller:     EntryCaller{true, 0, "github.com/cnotch/xlog/core_test.go", 30},
				Message:    "info message",
				Fields:     []Field{F("int", 100), F("str", "ok")},
				LoggerName: "",
				Ctx:        []Field{F("instance", 9000)},
			},
			InfoLevel.consoleString() + " 2019-01-18 12:00:35.000009 core_test.go:30: info message\n -  " + `{"instance":9000,"int":100,"str":"ok"}` + "\n",
		},
	}
	for _, tc := range cases {
		var buf bytes.Buffer
		core := NewCore(NewConsoleEncoder(LstdFlags|Lmicroseconds|Lshortfile), Lock(&buf), DebugLevel)
		core.Write(tc.e)
		s := string(buf.Bytes())
		if s != tc.want {
			t.Errorf("ioCore Out = \n%v, want = \n%v", s, tc.want)
		}
	}
}

func TestCore_Write_json(t *testing.T) {
	cases := []struct {
		e    Entry
		want string
	}{
		{
			Entry{
				Level:      InfoLevel,
				Time:       time.Date(2019, 1, 18, 12, 0, 35, 9876, time.UTC),
				Caller:     EntryCaller{true, 0, "github.com/cnotch/xlog/core_test.go", 30},
				Message:    "info message",
				Fields:     []Field{F("int", 100), F("str", "ok")},
				LoggerName: "",
				Ctx:        []Field{F("instance", 9000)},
			},
			`{"level":"INFO","time":"2019-01-18T12:00:35.000009876Z","caller":"core_test.go:30","msg":"info message","instance":9000,"int":100,"str":"ok"}` + "\n",
		},
	}
	for _, tc := range cases {
		var buf bytes.Buffer
		core := NewCore(NewJSONEncoder(Lshortfile), Lock(&buf), DebugLevel)
		core.Write(tc.e)
		s := string(buf.Bytes())
		if s != tc.want {
			t.Errorf("ioCore Out = \n%v, want = \n%v", s, tc.want)
		}
	}
}
