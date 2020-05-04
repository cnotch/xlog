// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"
)

const (
	// Tdate the date in the local time zone: 2006-01-02
	Tdate = 1 << iota
	// Ttimeprefix add 'T' between date and time if set, otherwise add space.
	// Only with Ttime is valid
	Ttimeprefix
	// Ttime the time in the local time zone: 15:04:05
	Ttime
	// Tmilliseconds the time in the local time zone: 15:04:05.000
	Tmilliseconds
	// Tmicroseconds the time in the local time zone: 15:04:05.000000
	Tmicroseconds
	// Tnanoseconds the time in the local time zone: 15:04:05.000000000
	Tnanoseconds
	// TnineFlag use with Tmilliseconds or Tmicroseconds or Tnanoseconds.
	// If use this flag, 000... pattern switch to 999...
	TnineFlag
	// Tzone the local time zone: Z07:00
	Tzone
	// Tdatetime the date and time in the local time zone: 2006-01-02 15:04:05"
	Tdatetime = Tdate | Ttime
	// TdatetimeMilli the date and time(ms) in the local time zone: 2006-01-02 15:04:05.000"
	TdatetimeMilli = Tdate | Ttime | Tmilliseconds
	// TdatetimeMicro the date and time(ms) in the local time zone: 2006-01-02 15:04:05.000000"
	TdatetimeMicro = Tdate | Ttime | Tmicroseconds
	// TdatetimeNano the date and time(ms) in the local time zone: 2006-01-02 15:04:05.000000000"
	TdatetimeNano = Tdate | Ttime | Tnanoseconds
	// Trfc3339 is equivalent to time.RFC3339: 2006-01-02T15:04:05Z07:00
	Trfc3339 = Tdate | Ttimeprefix | Ttime | Tzone
	// Trfc3339Nano is equivalent to time.RFC3339Nano: 2006-01-02T15:04:05.999999999Z07:00
	Trfc3339Nano = Tdate | Ttimeprefix | Ttime | Tnanoseconds | TnineFlag | Tzone
)

// Builder provides a convenient way to build strings using Write or Append methods.
// It minimizes memory copying. The zero value is ready to use.
// It implements io.Writer and io.ByteWriter and io.StringWriter.
type Builder struct {
	buf        []byte
	reflectEnc *json.Encoder // for encoding generic values by reflection
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (b *Builder) grow(n int) {
	buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
	copy(buf, b.buf)
	b.buf = buf
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
func (b *Builder) Grow(n int) {
	if n < 0 {
		panic("stringbuilder.Builder.Grow: negative count")
	}
	if cap(b.buf)-len(b.buf) < n {
		b.grow(n)
	}
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.buf = b.buf[:0]
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int {
	return len(b.buf)
}

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *Builder) Cap() int {
	return cap(b.buf)
}

// Bytes returns the builder's underlying byte slice.
func (b *Builder) Bytes() []byte {
	return b.buf
}

// String returns the accumulated string.
func (b *Builder) String() string {
	return *(*string)(unsafe.Pointer(b))
}

// Truncate discards all but the first n bytes from the b's buffer
// but continues to use the same allocated storage.
// It panics if n is negative or greater than b.Len().
func (b *Builder) Truncate(n int) {
	if n < 0 || n > b.Len() {
		panic("stringbuilder.Builder: truncation out of range")
	}
	b.buf = b.buf[:n]
}

// Write implements io.Writer, appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (b *Builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte implements io.ByteWriter, appends the byte c to b's buffer.
// The returned error is always nil.
func (b *Builder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
// It returns the length of r and a nil error.
func (b *Builder) WriteRune(r rune) (int, error) {
	if r < utf8.RuneSelf {
		b.buf = append(b.buf, byte(r))
		return 1, nil
	}

	l := len(b.buf)
	if cap(b.buf)-l < utf8.UTFMax {
		b.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(b.buf[l:l+utf8.UTFMax], r)
	b.buf = b.buf[:l+n]
	return n, nil
}

// WriteString implements io.StringWriter, appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// AppendBool appends "true" or "false", according to the value of v.
func (b *Builder) AppendBool(v bool) {
	b.buf = strconv.AppendBool(b.buf, v)
}

// AppendInt appends the string form of the int64 i.
func (b *Builder) AppendInt(i int64) {
	b.buf = strconv.AppendInt(b.buf, i, 10)
}

// AppendUint appends the string form of the uint64 i.
func (b *Builder) AppendUint(i uint64) {
	b.buf = strconv.AppendUint(b.buf, i, 10)
}

// AppendUintptr appends the string form of the uintptr p.
func (b *Builder) AppendUintptr(p uintptr) {
	b.WriteString("0x")
	b.buf = strconv.AppendUint(b.buf, uint64(p), 16)
}

// AppendFloat32 appends the string form of the float32 number f.
func (b *Builder) AppendFloat32(f float32) {
	b.appendFloat(float64(f), 32)
}

// AppendFloat64 appends the string form of the float64 number f.
func (b *Builder) AppendFloat64(f float64) {
	b.appendFloat(f, 64)
}

func (b *Builder) appendFloat(f float64, bits int) {
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}
	b.buf = strconv.AppendFloat(b.buf, f, fmt, -1, bits)
	if fmt == 'e' {
		s := b.buf
		// clean up e-09 to e-9
		n := len(s)
		if n >= 4 && s[n-4] == 'e' && s[n-3] == '-' && s[n-2] == '0' {
			s[n-2] = s[n-1]
			b.buf = s[:n-1]
		}
	}
}

// AppendComplex64 appends the string form of the complex64 number f.
func (b *Builder) AppendComplex64(val complex64) {
	r, i := real(val), imag(val)
	b.AppendFloat32(r)
	b.WriteByte('+')
	b.AppendFloat32(i)
	b.WriteByte('i')
}

// AppendComplex128 appends the string form of the complex128 number f.
func (b *Builder) AppendComplex128(val complex128) {
	r, i := real(val), imag(val)
	b.AppendFloat64(r)
	b.WriteByte('+')
	b.AppendFloat64(i)
	b.WriteByte('i')
}

// AppendDuration appends the string form of the time.Duration.String().
func (b *Builder) AppendDuration(d time.Duration) {
	// Largest time is 2540400h10m10.000000000s
	var buf [32]byte
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	if u < uint64(time.Second) {
		// Special case: if duration is smaller than a second,
		// use smaller units, like 1.2ms
		var prec int
		w--
		buf[w] = 's'
		w--
		switch {
		case u == 0:
			b.WriteString("0s")
		case u < uint64(time.Microsecond):
			// print nanoseconds
			prec = 0
			buf[w] = 'n'
		case u < uint64(time.Millisecond):
			// print microseconds
			prec = 3
			// U+00B5 'µ' micro sign == 0xC2 0xB5
			w-- // Need room for two bytes.
			copy(buf[w:], "µ")
		default:
			// print milliseconds
			prec = 6
			buf[w] = 'm'
		}
		w, u = fmtFrac(buf[:w], u, prec)
		w = fmtInt(buf[:w], u, -1)
	} else {
		w--
		buf[w] = 's'

		w, u = fmtFrac(buf[:w], u, 9)

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60, -1)
		u /= 60

		// u is now integer minutes
		if u > 0 {
			w--
			buf[w] = 'm'
			w = fmtInt(buf[:w], u%60, -1)
			u /= 60

			// u is now integer hours
			// Stop at hours because days can be different lengths.
			if u > 0 {
				w--
				buf[w] = 'h'
				w = fmtInt(buf[:w], u, -1)
			}
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}
	b.Write(buf[w:])
}

// AppendTime appends the textual representation in flag style to b.
// It has a faster formatting method that you can use if you are demanding
// performance, but it supports only a few formats
func (b *Builder) AppendTime(t time.Time, flag int) {
	// Largest time is 2006-01-02T15:04:05.999999999Z07:00
	var buf [40]byte
	w := len(buf)

	// zone +/-00:00
	if flag&Tzone != 0 {
		_, intervalSecond := t.Zone()
		if intervalSecond == 0 {
			w--
			buf[w] = 'Z'
		} else {
			prefix := byte('+')
			if intervalSecond < 0 {
				prefix = '-'
			}
			hour := intervalSecond / 3600
			min := (intervalSecond / 60) % 60

			w = fmtInt(buf[:w], uint64(min), 2)
			w--
			buf[w] = ':'
			w = fmtInt(buf[:w], uint64(hour), 2)
			w--
			buf[w] = prefix
		}
	}

	if flag&(Ttime|Tmilliseconds|Tmicroseconds|Tzone) != 0 {
		if flag&(Tmilliseconds|Tmicroseconds|Tnanoseconds) != 0 {
			if flag&Tnanoseconds != 0 {
				w = fmtNano(buf[:w], uint64(t.Nanosecond()), 9, flag&TnineFlag != 0)
			} else if flag&Tmicroseconds != 0 {
				w = fmtNano(buf[:w], uint64(t.Nanosecond()), 6, flag&TnineFlag != 0)
			} else {
				w = fmtNano(buf[:w], uint64(t.Nanosecond()), 3, flag&TnineFlag != 0)
			}
		}

		hour, min, sec := t.Clock()
		w = fmtInt(buf[:w], uint64(sec), 2)
		w--
		buf[w] = ':'
		w = fmtInt(buf[:w], uint64(min), 2)
		w--
		buf[w] = ':'
		w = fmtInt(buf[:w], uint64(hour), 2)

		// The prefix is required only if the date is valid
		if flag&Tdate != 0 {
			w--
			if flag&Ttimeprefix != 0 {
				buf[w] = 'T'
			} else {
				buf[w] = ' '
			}
		}
	}

	if flag&Tdate != 0 {
		year, month, day := t.Date()
		w = fmtInt(buf[:w], uint64(day), 2)
		w--
		buf[w] = '-'
		w = fmtInt(buf[:w], uint64(month), 2)
		w--
		buf[w] = '-'
		w = fmtInt(buf[:w], uint64(year), 4)
	}
	b.Write(buf[w:])
}

// AppendQuote appends a double-quoted Go string literal representing s.
func (b *Builder) AppendQuote(s string) {
	b.WriteByte('"')
	b.appendEscape(s, &safeSet)
	b.WriteByte('"')
}

// AppendHTMLQuote appends a double-quoted html string literal representing s.
func (b *Builder) AppendHTMLQuote(s string) {
	b.WriteByte('"')
	b.appendEscape(s, &htmlSafeSet)
	b.WriteByte('"')
}

// AppendByteSlice appends a base64 string representing []byte v.
func (b *Builder) AppendByteSlice(v []byte) {
	encodedLen := base64.StdEncoding.EncodedLen(len(v))
	b.Grow(encodedLen)
	dst := b.buf[b.Len() : b.Len()+encodedLen]
	base64.StdEncoding.Encode(dst, v)
	b.buf = b.buf[:b.Len()+encodedLen]
}

// AppendJSON appends an json-style string literal representing v.
// It implements a json-encoded subset of encoding/json and
// remains compatible with encoding/json.
func (b *Builder) AppendJSON(iv interface{}) (err error) {
	if iv == nil {
		b.WriteString("null")
		return
	}

	switch v := iv.(type) {
	case *string:
		b.AppendHTMLQuote(*v)
	case string:
		b.AppendHTMLQuote(v)
	case []string:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendHTMLQuote(e)
			}
			b.WriteByte(']')
		})
	case *bool:
		b.AppendBool(*v)
	case bool:
		b.AppendBool(v)
	case []bool:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendBool(e)
			}
			b.WriteByte(']')
		})
	case *int:
		b.AppendInt(int64(*v))
	case int:
		b.AppendInt(int64(v))
	case []int:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendInt(int64(e))
			}
			b.WriteByte(']')
		})
	case *int8:
		b.AppendInt(int64(*v))
	case int8:
		b.AppendInt(int64(v))
	case []int8:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendInt(int64(e))
			}
			b.WriteByte(']')
		})
	case *int16:
		b.AppendInt(int64(*v))
	case int16:
		b.AppendInt(int64(v))
	case []int16:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendInt(int64(e))
			}
			b.WriteByte(']')
		})
	case *int32:
		b.AppendInt(int64(*v))
	case int32:
		b.AppendInt(int64(v))
	case []int32:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendInt(int64(e))
			}
			b.WriteByte(']')
		})
	case *int64:
		b.AppendInt(int64(*v))
	case int64:
		b.AppendInt(int64(v))
	case []int64:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendInt(int64(e))
			}
			b.WriteByte(']')
		})
	case *uint:
		b.AppendUint(uint64(*v))
	case uint:
		b.AppendUint(uint64(v))
	case []uint:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendUint(uint64(e))
			}
			b.WriteByte(']')
		})
	case *uint8:
		b.AppendUint(uint64(*v))
	case uint8:
		b.AppendUint(uint64(v))
	case []uint8:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('"')
			b.AppendByteSlice(v)
			b.WriteByte('"')
		})
	case *uint16:
		b.AppendUint(uint64(*v))
	case uint16:
		b.AppendUint(uint64(v))
	case []uint16:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendUint(uint64(e))
			}
			b.WriteByte(']')
		})
	case *uint32:
		b.AppendUint(uint64(*v))
	case uint32:
		b.AppendUint(uint64(v))
	case []uint32:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendUint(uint64(e))
			}
			b.WriteByte(']')
		})
	case *uint64:
		b.AppendUint(uint64(*v))
	case uint64:
		b.AppendUint(uint64(v))
	case []uint64:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendUint(uint64(e))
			}
			b.WriteByte(']')
		})
	case uintptr:
		b.AppendUint(uint64(v))
	case unsafe.Pointer:
		b.AppendUintptr(uintptr(v))
	case *float32:
		b.AppendFloat32(*v)
	case float32:
		b.AppendFloat32(v)
	case []float32:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendFloat32(e)
			}
			b.WriteByte(']')
		})
	case *float64:
		b.AppendFloat64(*v)
	case float64:
		b.AppendFloat64(v)
	case []float64:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.AppendFloat64(e)
			}
			b.WriteByte(']')
		})
	case *complex64:
		b.WriteByte('"')
		b.AppendComplex64(*v)
		b.WriteByte('"')
	case complex64:
		b.WriteByte('"')
		b.AppendComplex64(v)
		b.WriteByte('"')
	case []complex64:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteByte('"')
				b.AppendComplex64(e)
				b.WriteByte('"')
			}
			b.WriteByte(']')
		})
	case *complex128:
		b.WriteByte('"')
		b.AppendComplex128(*v)
		b.WriteByte('"')
	case complex128:
		b.WriteByte('"')
		b.AppendComplex128(v)
		b.WriteByte('"')
	case []complex128:
		b.appendNullOrElse(v == nil, func() {
			b.WriteByte('[')
			for i, e := range v {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteByte('"')
				b.AppendComplex128(e)
				b.WriteByte('"')
			}
			b.WriteByte(']')
		})
	case *time.Duration:
		b.WriteByte('"')
		b.AppendDuration(*v)
		b.WriteByte('"')
	case time.Duration:
		b.WriteByte('"')
		b.AppendDuration(v)
		b.WriteByte('"')
	case *time.Time:
		b.WriteByte('"')
		b.AppendTime(*v, Trfc3339Nano)
		b.WriteByte('"')
	case time.Time:
		b.WriteByte('"')
		b.AppendTime(v, Trfc3339Nano)
		b.WriteByte('"')
	case error:
		b.AppendHTMLQuote(v.Error())
	default:
		len := b.Len()
		b.prepareReflectEnc()
		err = b.reflectEnc.Encode(v)
		if err != nil {
			b.buf = b.buf[:len]
			return
		}

		// ignore json.Encoder last '\n'
		b.buf = b.buf[:b.Len()-1]
	}
	return
}

func (b *Builder) prepareReflectEnc() {
	if b.reflectEnc == nil {
		b.reflectEnc = json.NewEncoder(b)
		// b.reflectEnc.SetIndent("", b.indent)
	}
}

func (b *Builder) appendNullOrElse(isNil bool, elseOp func()) {
	if isNil {
		b.WriteString("null")
	} else {
		elseOp()
	}
}

// For JSON-escaping
const _hex = "0123456789abcdef"

// safeSet holds the value true if the ASCII character with the given array
// position can be represented inside a JSON string without any further
// escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), and the backslash character ("\").
var safeSet [utf8.RuneSelf]bool

// htmlSafeSet holds the value true if the ASCII character with the given
// array position can be safely represented inside a JSON string, embedded
// inside of HTML <script> tags, without any additional escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), the backslash character ("\"), HTML opening and closing
// tags ("<" and ">"), and the ampersand ("&").
var htmlSafeSet [utf8.RuneSelf]bool
var htmlCompactSafeSet [utf8.RuneSelf]bool

func init() {
	for i := byte(0x20); i < utf8.RuneSelf; i++ {
		safeSet[i] = true
		htmlSafeSet[i] = true
		htmlCompactSafeSet[i] = true
	}
	safeSet['\\'] = false
	safeSet['"'] = false

	htmlSafeSet['\\'] = false
	htmlSafeSet['"'] = false
	htmlSafeSet['<'] = false
	htmlSafeSet['>'] = false
	htmlSafeSet['&'] = false
	htmlCompactSafeSet['<'] = false
	htmlCompactSafeSet['>'] = false
	htmlCompactSafeSet['&'] = false
}

func (b *Builder) appendEscape(s string, safeSet *[utf8.RuneSelf]bool) {
	start := 0
	for i := 0; i < len(s); {
		if c := s[i]; c < utf8.RuneSelf {
			if (*safeSet)[c] {
				i++
				continue
			}
			if start < i { // safe char
				b.WriteString(s[start:i])
			}
			b.WriteByte('\\') // add escape character
			switch c {
			case '\\', '"':
				b.WriteByte(c)
			case '\n':
				b.WriteByte('n')
			case '\r':
				b.WriteByte('r')
			case '\t':
				b.WriteByte('t')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				b.WriteString(`u00`)
				b.WriteByte(_hex[c>>4])
				b.WriteByte(_hex[c&0xF])
			}
			i++
			start = i // reset next start
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			if start < i {
				b.WriteString(s[start:i])
			}
			b.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}

		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if r == '\u2028' || r == '\u2029' {
			if start < i {
				b.WriteString(s[start:i])
			}
			b.WriteString(`\u202`)
			b.WriteByte(_hex[r&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
}

// fmtFrac formats the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf, omitting trailing zeros. It omits the decimal
// point too when the fraction is 0. It returns the index where the
// output bytes begin and the value v/10**prec.
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	print := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		print = print || digit != 0
		if print {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if print {
		w--
		buf[w] = '.'
	}
	return w, v
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func fmtInt(buf []byte, i uint64, wid int) int {
	w := len(buf)
	for i >= 10 || wid > 1 {
		wid--
		w--
		buf[w] = byte(i%10) + '0'
		i /= 10
	}
	// i < 10
	w--
	buf[w] = byte('0' + i)
	return w
}

func fmtNano(buf []byte, nano uint64, wid int, trimZeroSuffix bool) int {
	w := len(buf)
	m := 9 - wid
	for m > 0 {
		nano /= 10
		m--
	}

	if nano > 0 || !trimZeroSuffix {
		w = fmtInt(buf[:w], nano, wid)

		if trimZeroSuffix { // trim '0' suffix
			l := len(buf)
			for buf[l-1] == '0' {
				l--
			}

			zeroNum := len(buf) - l
			if zeroNum > 0 {
				copy(buf[w+zeroNum:], buf[w:])
				w += zeroNum
			}
		}

		w--
		buf[w] = '.'
	}

	return w
}

var builderPool = sync.Pool{
	New: func() interface{} {
		return &Builder{buf: make([]byte, 0, 512)}
	},
}

func getBuilder() *Builder {
	b := builderPool.Get().(*Builder)
	b.Reset()
	return b
}

func putBuilder(b *Builder) {
	builderPool.Put(b)
}
