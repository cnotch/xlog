// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package xlog

import (
	"io"
	"sync"
)

// Lock wraps a io.Writer in a mutex to make it safe for concurrent use.
// In particular, *os.Files must be locked before use.
func Lock(w io.Writer) io.Writer {
	if _, ok := w.(*lockedWriter); ok {
		// no need to layer on another lock
		return w
	}
	return &lockedWriter{w: w, sync: getSyncFunc(w)}
}

type lockedWriter struct {
	sync.Mutex
	w    io.Writer
	sync func() error
}

func (s *lockedWriter) Write(bs []byte) (int, error) {
	s.Lock()
	n, err := s.w.Write(bs)
	s.Unlock()
	return n, err
}

func (s *lockedWriter) Sync() error {
	if s.sync == nil {
		return nil
	}

	s.Lock()
	err := s.sync()
	s.Unlock()
	return err
}

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
//
// Each write is written to each listed writer, one at a time.
// If a listed writer returns an error, that overall write operation
// stops and returns the error; it does not continue down the list.
func MultiWriter(writers ...io.Writer) io.Writer {
	allWriters := make([]io.Writer, 0, len(writers))
	allSyncs := make([]func() error, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
			allSyncs = append(allSyncs, mw.syncs...)
		} else {
			allWriters = append(allWriters, w)
			sync := getSyncFunc(w)
			if sync != nil {
				allSyncs = append(allSyncs, sync)
			}
		}
	}
	return &multiWriter{allWriters, allSyncs}
}

type multiWriter struct {
	writers []io.Writer
	syncs   []func() error
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		wn, werr := w.Write(p)
		if werr != nil {
			err = combineErrors(err, werr)
			continue
		}

		if wn != len(p) {
			err = combineErrors(err, io.ErrShortWrite)
		}
		if wn > n {
			n = wn
		}
	}
	return
}

func (mw *multiWriter) Sync() (err error) {
	for _, sync := range mw.syncs {
		err = combineErrors(err, sync())
	}
	return
}

type syncer interface {
	Sync() error
}

type flusher interface {
	Flush() error
}

// Get the known Sync function
func getSyncFunc(w io.Writer) func() error {
	switch w := w.(type) {
	case syncer:
		return w.Sync
	case flusher:
		return w.Flush
	default:
		return nil
	}
}
