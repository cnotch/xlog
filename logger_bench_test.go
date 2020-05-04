// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

type user struct {
	Name      string
	Email     string
	CreatedAt time.Time
}

var _jane = &user{
	Name:      "Jane Doe",
	Email:     "jane@test.com",
	CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
}

func BenchmarkCallerf(b *testing.B) {
	withBenchedLoggerWithCaller(b, func(log *Logger) {
		log.Infof("Callerf. %f", 56.67)
	})
}

func BenchmarkNoCaller(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("No context.")
	})
}

func BenchmarkBoolField(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Boolean.", F("foo", true))
	})
}

func BenchmarkIntField(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Integer.", F("foo", 42))
	})
}

func BenchmarkFloat64Field(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Floating point.", F("foo", 3.14))
	})
}
func BenchmarkStringField(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Strings.", F("foo", "bar"))
	})
}

func BenchmarkTimeField(b *testing.B) {
	t := time.Unix(0, 0)
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Time.", F("foo", t))
	})
}

func BenchmarkDurationField(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Duration", F("foo", time.Second))
	})
}

func BenchmarkErrorField(b *testing.B) {
	err := errors.New("egad")
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Error.", F("error", err))
	})
}

func BenchmarkReflectField(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Reflection-based serialization.", F("user", _jane))
	})
}

func BenchmarkIntsField(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Errors.", F("ints", []int{100, 200, 300, 400}))
	})
}
func BenchmarkAddCallerHook(b *testing.B) {
	withBenchedLoggerWithCaller(b, func(log *Logger) {
		log.Info("Caller.")
	})
}

func Benchmark10Fields(b *testing.B) {
	withBenchedLogger(b, func(log *Logger) {
		log.Info("Ten fields, passed at the log site.",
			F("one", 1),
			F("two", 2),
			F("three", 3),
			F("four", 4),
			F("five", 5),
			F("six", 6),
			F("seven", 7),
			F("eight", 8),
			F("nine", 9),
			F("ten", 10),
		)
	})
}

func Benchmark100Fields(b *testing.B) {
	const batchSize = 50
	logger := New(
		NewCore(NewJSONEncoder(0), (ioutil.Discard), DebugLevel))

	// Don't include allocating these helper slices in the benchmark. Since
	// access to them isn't synchronized, we can't run the benchmark in
	// parallel.
	first := make([]Field, batchSize)
	second := make([]Field, batchSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i := 0; i < batchSize; i++ {
			// We're duplicating keys, but that doesn't affect performance.
			first[i] = F("foo", i)
			second[i] = F("foo", i+batchSize)
		}
		logger.With(Fields(first...)).Info("Child loggers with lots of context.", second...)
		// logger.With(first...).Info("Child loggers with lots of context.")
		// logger.Info("Child loggers with lots of context.", second...)
	}
}

func withBenchedLogger(b *testing.B, f func(*Logger)) {
	logger := New(
		NewCore(NewJSONEncoder(0), (ioutil.Discard), DebugLevel))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func withBenchedLoggerWithCaller(b *testing.B, f func(*Logger)) {
	logger := New(
		NewCore(NewConsoleEncoder(LstdFlags|Lmicroseconds|Llongfile), (ioutil.Discard), DebugLevel),
		AddCaller())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func withBenchedStdLogger(b *testing.B, f func(*log.Logger)) {
	logger := log.New(ioutil.Discard, "benchTest", log.LstdFlags|log.Lmicroseconds|log.Llongfile)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkStdLoggerWithCaller(b *testing.B) {
	withBenchedStdLogger(b, func(log *log.Logger) {
		log.Output(1, "Caller.")
	})
}
