// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

// These flags define which text to prefix to each log entry generated by the Logger.
// Bits are or'ed together to control what's printed.
// There is no control over the order they appear (the order listed
// here) or the format they present (as described in the comments).
// The prefix is followed by a colon only when Llongfile or Lshortfile
// is specified.
// For example, flags Ldate | Ltime (or LstdFlags) produce,
//	2009-01-23 01:23:23 message
// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
//	2009-01-23 01:23:23.123123 /a/b/c/d.go:23: message
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009-01-23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// Encoder is a format-agnostic interface for all log entry marshalers.
type Encoder interface {
	// Encode encodes a log entry to b
	Encode(b *Builder, e Entry) error
}

// NewConsoleEncoder returns an encoder whose output is designed for human -
// rather than machine - consumption.
func NewConsoleEncoder(flags int) Encoder { return consoleEncoder(flags) }

// NewJSONEncoder returns a fast, low-allocation JSON encoder.
// The encoder appropriately escapes all field keys and values.
func NewJSONEncoder(flags int) Encoder { return jsonEncoder(flags) }

type consoleEncoder int

func (enc consoleEncoder) Encode(b *Builder, e Entry) error {
	flags := int(enc)
	// Level
	b.WriteString(e.Level.consoleString())
	// Time
	if tflag := timeFlags(flags); tflag != 0 {
		t := e.Time
		if flags&LUTC != 0 {
			t = t.UTC()
		}
		b.WriteByte(' ')
		b.AppendTime(t, tflag)
		b.WriteByte(' ')
	} else {
		b.WriteByte(' ')
	}

	// Name
	i := 0
	if e.LoggerName != "" {
		b.WriteString(e.LoggerName)
		i++
	}

	// Caller
	if flags&(Llongfile|Lshortfile) != 0 && e.Caller.Defined {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteString(callerFile(e.Caller.File, flags))
		b.WriteByte(':')
		b.AppendInt(int64(e.Caller.Line))
		i++
	}

	// Message
	if i > 0 {
		b.WriteString(": ")
	}
	b.WriteString(e.Message)
	b.WriteByte('\n')

	// Fields
	if len(e.Ctx) > 0 || len(e.Fields) > 0 {
		i = 0
		b.WriteString(" -  ")
		b.WriteByte('{')
		if len(e.Ctx) > 0 {
			O(e.Ctx).appendTo(b)
			i += len(e.Ctx)
		}
		if len(e.Fields) > 0 {
			if i > 0 {
				b.WriteByte(',')
			}
			O(e.Fields).appendTo(b)
		}
		b.WriteString("}\n")
	}
	return nil
}

type jsonEncoder int

func (enc jsonEncoder) Encode(b *Builder, e Entry) error {
	flags := int(enc)
	b.WriteByte('{')

	b.WriteString(`"level":"`)
	b.WriteString(e.Level.CapitalString())
	b.WriteByte('"')

	b.WriteString(`,"time":`)
	b.WriteByte('"')
	b.AppendTime(e.Time, Trfc3339Nano)
	b.WriteByte('"')

	if e.LoggerName != "" {
		b.WriteString(`,"logger":`)
		b.AppendHTMLQuote(e.LoggerName)
	}

	if flags&(Llongfile|Lshortfile) != 0 && e.Caller.Defined {
		b.WriteString(`,"caller":"`)
		b.WriteString(callerFile(e.Caller.File, flags))
		b.WriteByte(':')
		b.AppendInt(int64(e.Caller.Line))
		b.WriteByte('"')
	}

	b.WriteString(`,"msg":`)
	b.AppendHTMLQuote(e.Message)

	if len(e.Ctx) > 0 {
		b.WriteByte(',')
		O(e.Ctx).appendTo(b)
	}
	if len(e.Fields) > 0 {
		b.WriteByte(',')
		O(e.Fields).appendTo(b)
	}
	b.WriteString("}\n")
	return nil
}

func timeFlags(flags int) int {
	tflag := 0
	if flags&Ldate != 0 {
		tflag |= Tdate
	}
	if flags&Ltime != 0 {
		tflag |= Ttime
	}
	if flags&Lmicroseconds != 0 {
		tflag |= Tmicroseconds
	}
	return tflag
}

func callerFile(file string, flags int) string {
	if flags&Lshortfile != 0 {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	}
	return file
}
