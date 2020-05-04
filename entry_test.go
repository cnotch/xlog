// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"encoding/base64"
	"encoding/json"
	"runtime"
	"testing"
	"time"
)

func TestNewEntryCaller(t *testing.T) {
	wantFile := "github.com/cnotch/xlog/entry_test.go"
	caller := NewEntryCaller(runtime.Caller(0))
	if caller.File != wantFile {
		t.Errorf("NewEntryCaller() File = %v, want %v", caller.File, wantFile)
	}
}

func TestField_String(t *testing.T) {
	var _jane = &struct {
		Name      string
		Email     string
		CreatedAt time.Time
	}{
		Name:      "Jane Doe",
		Email:     "jane@test.com",
		CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jane, _ := json.Marshal(_jane)
	data := []byte("any + old & data")
	str := base64.StdEncoding.EncodeToString(data)
	tm := time.Now()
	// time.Date(1980,1,1,12,0,0,0,time.UTC),
	var testCases = []struct {
		name string
		f    Field
		want string
	}{
		{
			"Binary",
			F("binary", []byte{0x45, 0x56, 0x99, 0xf8, 0xff, 0x00}),
			`"binary":"` + base64.StdEncoding.EncodeToString([]byte{0x45, 0x56, 0x99, 0xf8, 0xff, 0x00}) + `"`,
		},
		{
			"String",
			F("str", "ok"),
			`"str":"ok"`,
		},
		{
			"Duration",
			F("interval", time.Duration(99988834500)),
			`"interval":"` + time.Duration(99988834500).String() + `"`,
		},
		{
			"Time",
			F("interval", tm),
			`"interval":"` + tm.Format(time.RFC3339Nano) + `"`,
		},
		{
			"Bool",
			F("bool", true),
			`"bool":true`,
		},
		{
			"Complex128",
			F("cmpl128", 3.1+4.2i),
			`"cmpl128":"3.1+4.2i"`,
		},
		{
			"Complex64",
			F("cmpl64", complex64(3.1+4.2i)),
			`"cmpl64":"3.1+4.2i"`,
		},
		{
			"Float64",
			F("f64", 3.142),
			`"f64":3.142`,
		},
		{
			"Float32",
			F("f32", float32(3.142)),
			`"f32":3.142`,
		},
		{
			"Int",
			F("i", 123),
			`"i":123`,
		},
		{
			"Uint",
			F("i", uint(123)),
			`"i":123`,
		},
		{
			"Reflect",
			F("reflect", _jane),
			`"reflect":` + string(jane),
		},
		{
			"Binary",
			F("binary", data),
			`"binary":"` + str + `"`,
		},
		{
			"Strings",
			F("ss", []string{"cao", "jia", "hong"}),
			`"ss":["cao","jia","hong"]`,
		},
		{
			"Bools",
			F("ss", []bool{true, false, true}),
			`"ss":[true,false,true]`,
		},
		{
			"Complex128s",
			F("ss", []complex128{1.1 + 1.1i, 2.2 + 2.2i, 3.3 + 3.3i, 4.4 + 4.4i}),
			`"ss":["1.1+1.1i","2.2+2.2i","3.3+3.3i","4.4+4.4i"]`,
		},
		{
			"Float32s",
			F("ss", []float32{1.1, 2.2, 3.3, 4.4}),
			`"ss":[1.1,2.2,3.3,4.4]`,
		},
		{
			"Ints",
			F("ss", []int{1, 2, 3, 4}),
			`"ss":[1,2,3,4]`,
		},
		{
			"Uints",
			F("ss", []uint{1, 2, 3, 4}),
			`"ss":[1,2,3,4]`,
		},
		{
			"Object",
			F("obj", O{F("name", "chj")}),
			`"obj":{"name":"chj"}`,
		},
		{
			"Object",
			F("obj", O{F("name", "chj"), F("age", 45)}),
			`"obj":{"name":"chj","age":45}`,
		},
		{
			"Array",
			F("ss", []O{{F("name", "chj"), F("age", 45)},
				{F("name", "chj2"), F("age", 30)}}),
			`"ss":[{"name":"chj","age":45},{"name":"chj2","age":30}]`,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("%s() = %v,want %v", tt.name, got, tt.want)
			}
		})
	}
}
