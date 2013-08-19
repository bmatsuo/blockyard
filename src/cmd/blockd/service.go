// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// service.go [created: Sun, 18 Aug 2013]

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

var ErrDigestMismatch = errors.New("digest did not match content")
var ErrTooLarge = fmt.Errorf("too large")
var ErrUnexpectedEOF = fmt.Errorf("unexpected eof")

type API interface {
	Get(id string) (Block, error)
	Create(r io.Reader, checksum string) (string, error)
	Delete(id string) error
}

type Service struct {
}

func NewService() *Service {
	s := new(Service)
	return s
}

func (service *Service) Get(id string) (Block, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (service *Service) Create(r io.Reader, size uint64, digest string) (string, error) {
	// read request data and fork to file system and checksum
	hash := sha1.New()
	tee := newTWriter(hash, ioutil.Discard) // FIXME
	err := copyN(tee, r, size)
	if err != nil {
		return "", err
	}

	computedDigest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	if digest != computedDigest {
		return "", ErrDigestMismatch
	}

	return digest, nil
}

func (service *Service) Delete(id string) error {
	return fmt.Errorf("unimplemented")
}

type Block interface {
	Id() string
	ModTime() time.Time
	Size() uint64
	Open() (io.ReadCloser, error)
}

type simpleBlock struct {
	id   string
	data []byte
}

func (b *simpleBlock) Id() string {
	return b.id
}

func (b *simpleBlock) ModeTime() time.Time {
	return time.Now()
}

func (b *simpleBlock) Size() uint64 {
	return uint64(len(b.data))
}

func (b *simpleBlock) Open() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer(b.data))
}

func copyN(w io.Writer, r io.Reader, n uint64) error {
	bufSize := 4096
	if n%uint64(bufSize) == 0 {
		bufSize++
	}

	var totalread uint64
	buf := make([]byte, bufSize)
	for {
		nread, err := r.Read(buf)
		if err == io.EOF {
			if nread == 0 {
				break
			}
		} else if err != nil {
			return err
		}

		_nwrite := nread
		if totalread+uint64(nread) > n {
			// pushed over just now.
			_nwrite = int(n - totalread)
		}
		totalread += uint64(nread)

		_, err = w.Write(buf[:_nwrite])
		if err != nil {
			return err
		}

		if totalread > n {
			return ErrTooLarge
		}
	}

	if totalread < n {
		return ErrUnexpectedEOF
	}

	return nil
}

type writeResponse struct {
	id  int
	n   int
	err error
}

type tWriter struct {
	resp       chan writeResponse
	out1, out2 io.Writer
}

func newTWriter(out1, out2 io.Writer) *tWriter {
	return &tWriter{
		resp: make(chan writeResponse, 0),
		out1: out1,
		out2: out2,
	}
}

func (w tWriter) Write(p []byte) (n int, err error) {
	go func() {
		n, err := w.out1.Write(p)
		w.resp <- writeResponse{1, n, err}
	}()
	go func() {
		n, err := w.out2.Write(p)
		w.resp <- writeResponse{1, n, err}
	}()
	resp1 := <-w.resp
	resp2 := <-w.resp
	if resp1.err != nil {
		return resp1.n, resp1.err
	}
	if resp2.err != nil {
		return resp2.n, resp2.err
	}
	return len(p), nil
}
