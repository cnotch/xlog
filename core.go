// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"io"
)

// Core is a minimal, fast logger interface.
// It's designed for library authors to wrap in a more user-friendly API.
type Core interface {
	LevelEnabler
	// Write serializes the Entry supplied at the log site and
	// writes them to their destination.
	//
	// If called, Write should always log the Entry; it should not
	// replicate the logic of Check.
	Write(e Entry) error
	// Sync flushes buffered logs (if any).
	Sync() error
}

type nopCore struct{}

// NewNopCore returns a no-op Core.
func NewNopCore() Core                 { return nopCore{} }
func (nopCore) Enabled(lvl Level) bool { return false }
func (nopCore) Write(e Entry) error    { return nil }
func (nopCore) Sync() error            { return nil }

type ioCore struct {
	enc          Encoder
	w            io.Writer // destination for output
	LevelEnabler           // available log levels
	sync         func() error
}

// NewCore creates a Core that writes logs to a io.Writer.
func NewCore(enc Encoder, w io.Writer, enab LevelEnabler) Core {
	c := &ioCore{
		enc:          enc,
		LevelEnabler: enab,
		w:            w,
	}
	c.sync = getSyncFunc(w)
	return c
}

func (c *ioCore) Write(e Entry) (err error) {
	b := getBuilder()
	defer putBuilder(b)

	if err = c.enc.Encode(b, e); err == nil {
		_, err = c.w.Write(b.Bytes())
	}

	if err == nil && e.Level >= ErrorLevel {
		err = c.Sync()
	}
	return
}

func (c *ioCore) Sync() error {
	if c.sync != nil {
		return c.sync()
	}
	return nil
}

type multiCore struct {
	cores         []Core
	levelsEnabled [_maxLevel + 2]bool
}

// NewTee creates a Core that duplicates log entries into two or more
// underlying Cores.
//
// Calling it with a single Core returns the input unchanged, and calling
// it with no input returns a no-op Core.
func NewTee(cores ...Core) Core {
	switch len(cores) {
	case 0:
		return nopCore{}
	case 1:
		return cores[0]
	}

	allCores := make([]Core, 0, len(cores))
	for _, c := range cores {
		if mc, ok := c.(*multiCore); ok {
			allCores = append(allCores, mc.cores...)
		} else {
			allCores = append(allCores, c)
		}
	}

	var levelsEnabled [_maxLevel + 2]bool
	for _, c := range allCores {
		for lvl := _minLevel; lvl < _maxLevel+1; lvl++ {
			if c.Enabled(lvl) {
				levelsEnabled[lvl+1] = true
			}
		}
	}
	return &multiCore{allCores, levelsEnabled}
}

func (mc *multiCore) Enabled(lvl Level) bool {
	if lvl < _minLevel || lvl > _maxLevel {
		return false
	}
	return mc.levelsEnabled[lvl+1]
}

func (mc *multiCore) Write(e Entry) (err error) {
	for _, c := range mc.cores {
		cerr := c.Write(e)
		if cerr != nil {
			err = combineErrors(err, cerr)
		}
	}
	return
}

func (mc *multiCore) Sync() (err error) {
	for _, c := range mc.cores {
		cerr := c.Sync()
		if cerr != nil {
			err = combineErrors(err, cerr)
		}
	}
	return
}
