// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

var formatTimeTestCases = []struct {
	layout string
	flag   int
}{
	{"2006-01-02", Tdate},
	{"15:04:05", Ttime},
	{"2006-01-02 15:04:05", Tdatetime},
	{"2006-01-02T15:04:05", Tdatetime | Ttimeprefix},
	{"2006-01-02 15:04:05.000", TdatetimeMilli},
	{"2006-01-02 15:04:05.000000", TdatetimeMicro},
	{"2006-01-02 15:04:05.000000000", TdatetimeNano},
	{"2006-01-02 15:04:05.999", TdatetimeMilli | TnineFlag},
	{"2006-01-02 15:04:05.999999", TdatetimeMicro | TnineFlag},
	{"2006-01-02 15:04:05.999999999", TdatetimeNano | TnineFlag},
	{time.RFC3339, Trfc3339},
	{time.RFC3339Nano, Trfc3339Nano},
}

func TestBuilder_AppendTime(t *testing.T) {
	times := []time.Time{
		time.Now(),
		time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(1980, 1, 1, 12, 0, 0, 0, time.Now().Location()),
		time.Date(1980, 1, 1, 12, 0, 0, 1234, time.UTC),
		time.Date(1980, 1, 1, 12, 0, 0, 1234, time.Now().Location()),
		time.Date(1980, 1, 1, 12, 0, 0, 123456789, time.Now().Location()),
		time.Date(2019, 1, 18, 12, 0, 35, 9876, time.UTC),
	}
	for _, tm := range times {
		for _, tt := range formatTimeTestCases {
			t.Run("builder.AppendTime("+tt.layout+")", func(t *testing.T) {
				want := tm.Format(tt.layout)

				var builder Builder
				builder.AppendTime(tm, tt.flag)
				got := builder.String()
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s = %v, want %v", tt.layout, got, want)
				}
			})
		}
	}
}

func TestBuilder_AppendDuration(t *testing.T) {
	t.Run("builder.AppendDuration", func(t *testing.T) {
		d := time.Duration(91989993334522)
		want := d.String()
		var builder Builder
		builder.AppendDuration(d)
		got := builder.String()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Builder.AppendDuration() = %v, want %v", got, want)
		}
	})
}

func TestBuilder_AppendQuote(t *testing.T) {
	testStrs := []string{
		`"Fran & Freddie's Diner"`,
		`\t\n\r\\`,
		"中文就是好\n",
	}
	for _, s := range testStrs {
		t.Run("builder.AppendQuote", func(t *testing.T) {
			want := strconv.Quote(s)
			var builder Builder
			builder.AppendQuote(s)
			got := builder.String()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Builder.AppendQuote = %v, want %v", got, want)
			}
		})
	}
}

func TestBuild_AppendJSON(t *testing.T) {
	type Embed struct {
		F float64
	}

	type embed struct {
		Year int
	}

	str := "strtest\n"
	bv := true
	iv, i8v, i16v, i32v, i64v := int(100), int8(100), int16(100), int32(100), int64(100)
	uiv, ui8v, ui16v, ui32v, ui64v := uint(100), uint8(100), uint16(100), uint32(100), uint64(100)
	f32v, f64v := float32(3.2), 6.4
	comp64v, comp128v := complex64(complex(3.2, 3.3)), complex(6.4, 6.5)
	dt := time.Duration(1234567)
	tm := time.Now()

	embptr := &Embed{9.9}

	tests := []struct {
		label string
		input interface{}
		want  string
	}{
		{"nil", nil, "null"},
		{"string", str, `"strtest\n"`},
		{"*string", &str, `"strtest\n"`},
		{"[]string", []string{str, str}, `["strtest\n","strtest\n"]`},
		{"bool", bv, "true"},
		{"*bool", &bv, "true"},
		{"[]bool", []bool{true, false, true}, "[true,false,true]"},
		{"int", iv, "100"},
		{"*int", &iv, "100"},
		{"[]int", []int{100, 110, 120}, "[100,110,120]"},
		{"int8", i8v, "100"},
		{"*int8", &i8v, "100"},
		{"[]int8", []int8{100, 110, 120}, "[100,110,120]"},
		{"int16", i16v, "100"},
		{"*int16", &i16v, "100"},
		{"[]int16", []int16{100, 110, 120}, "[100,110,120]"},
		{"int32", i32v, "100"},
		{"*int32", &i32v, "100"},
		{"[]int32", []int32{100, 110, 120}, "[100,110,120]"},
		{"int64", i64v, "100"},
		{"*int64", &i64v, "100"},
		{"[]int64", []int64{100, 110, 120}, "[100,110,120]"},
		{"uint", uiv, "100"},
		{"*uint", &uiv, "100"},
		{"[]uint", []uint{100, 110, 120}, "[100,110,120]"},
		{"uint8", ui8v, "100"},
		{"*uint8", &ui8v, "100"},
		{"[]uint8", []uint8{100, 110, 120}, `"ZG54"`},
		{"uint16", ui16v, "100"},
		{"*uint16", &ui16v, "100"},
		{"[]uint16", []uint16{100, 110, 120}, "[100,110,120]"},
		{"uint32", ui32v, "100"},
		{"*uint32", &ui32v, "100"},
		{"[]uint32", []uint32{100, 110, 120}, "[100,110,120]"},
		{"uint64", ui64v, "100"},
		{"*uint64", &ui64v, "100"},
		{"[]uint64", []uint64{100, 110, 120}, "[100,110,120]"},
		{"float32", f32v, "3.2"},
		{"*float32", &f32v, "3.2"},
		{"[]float32", []float32{100, 110, 120}, "[100,110,120]"},
		{"float64", f64v, "6.4"},
		{"*float64", &f64v, "6.4"},
		{"[]float64", []float64{100, 110, 120}, "[100,110,120]"},
		{"complex64", comp64v, `"3.2+3.3i"`},
		{"*complex64", &comp64v, `"3.2+3.3i"`},
		{"[]complex64", []complex64{comp64v, comp64v, comp64v}, `["3.2+3.3i","3.2+3.3i","3.2+3.3i"]`},
		{"complex128", comp128v, `"6.4+6.5i"`},
		{"*complex128", &comp128v, `"6.4+6.5i"`},
		{"[]complex128", []complex128{comp128v, comp128v, comp128v}, `["6.4+6.5i","6.4+6.5i","6.4+6.5i"]`},
		{"duration", dt, `"` + dt.String() + `"`},
		{"*duration", &dt, `"` + dt.String() + `"`},
		{"time", tm, `"` + tm.Format(time.RFC3339Nano) + `"`},
		{"*time", &tm, `"` + tm.Format(time.RFC3339Nano) + `"`},
		{"struct(embed)", struct {
			Name string
			Age  int
			Embed
		}{"chj", 40, Embed{1.1}}, `{"Name":"chj","Age":40,"F":1.1}`},
		{"struct(embed ptr)", struct {
			Name string
			Age  int
			*Embed
		}{"chj", 40, &Embed{1.1}}, `{"Name":"chj","Age":40,"F":1.1}`},
		{"struct(tag)", struct {
			Name  string    `json:"name,omitempty"`
			Age   int       `json:",omitempty"`
			Birth time.Time `json:"-"`
			Embed `json:"emb"`
			Emb2  **Embed
			embed
			Son embed
			F64 float64 `json:",string"`
		}{"chj", 0, time.Now(), Embed{1.1}, &embptr, embed{45}, embed{17}, 2.1}, `{"name":"chj","emb":{"F":1.1},"Emb2":{"F":9.9},"Year":45,"Son":{"Year":17},"F64":"2.1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			var b Builder
			if err := b.AppendJSON(tt.input); err != nil {
				t.Errorf("Builder.AppendJSON() error = %v", err)
			} else {
				got := b.String()
				if got != tt.want {
					t.Errorf("Builder.AppendJSON = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
func BenchmarkStd_AppendTime(b *testing.B) {
	var sb Builder
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.buf = now.AppendFormat(sb.buf, time.RFC3339Nano)
	}
}

func BenchmarkBuild_AppendTime(b *testing.B) {
	var sb Builder
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.AppendTime(now, Trfc3339Nano)
	}
}

func BenchmarkStd_Quote(b *testing.B) {
	var sb Builder
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.buf = strconv.AppendQuote(sb.buf, "builder provides a convenient way to build strings.\n")
	}
}

func BenchmarkBuilder_Quote(b *testing.B) {
	var sb Builder
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.AppendQuote("builder provides a convenient way to build strings.\n")
	}
}

func BenchmarkBuilder_HTMLQuote(b *testing.B) {
	var sb Builder
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sb.Reset()
		sb.AppendHTMLQuote("builder provides a convenient way to build strings.\n")
	}
}
