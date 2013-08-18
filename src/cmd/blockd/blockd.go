// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// blockd.go [created: Wed, 14 Aug 2013]

/*
A REST block storage service. Blockd is designed to be a node in a distributed
file system. The blocks it store are parts of files. But, the entire file is
rarely stored entirely on one node. Blockd nodes are ignorant about the presence
of any other blockd nodes.
*/
package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/bmatsuo/go-syslog"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"schnutil/log"
	"schnutil/stat"
	"strconv"
	"syscall"
	"time"
)

func main() {
	log.SyslogBase = "blockd"
	logger, err := log.NewSyslog(syslog.LOG_NOTICE, "")
	if err != nil {
		panic(err)
	}

	defer func() {
		if e := recover(); e != nil {
			logger.Crit("unhandled runtime panic: " + fmt.Sprint(e))
			callers := make([]uintptr, 20)
			n := runtime.Callers(1, callers)
			for i := 0; i < n; i++ {
				pc := callers[i]
				fn := runtime.FuncForPC(pc)
				if fn == nil {
					logger.Crit(fmt.Sprintf("[i] unknown FuncForPC: %v", i, pc))
				} else {
					name := fn.Name()
					file, line := fn.FileLine(pc)
					offset := pc - fn.Entry()
					frame := fmt.Sprintf("[%d] %s (%s:%v) +0x%X",
						i, name, file, line, offset)
					logger.Notice(frame)
				}
			}
			os.Exit(1)
		} else {
			logger.Notice("shut down complete")
			os.Exit(0)
		}
	}()

	statdaemon := stat.NewRuntimeStatDaemon(30 * time.Second)
	err = statdaemon.Start()
	if err != nil {
		panic(err)
	}

	logger.Notice(fmt.Sprint("serving HTTP at ", ":8080"))
	httpserver, err := NewHTTPServerAddr(":8080")
	if err != nil {
		panic(err)
	}
	httpserver.Handler = Routes()
	httpErrorLogger, err := log.NewSyslog(syslog.LOG_NOTICE, "http.server_error")
	if err != nil {
		panic(err)
	}

	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	httpdone := make(chan error, 1)
	defer func() {
		for err := range httpdone {
			if err != nil {
				logger.Crit(err.Error())
				os.Exit(1)
			}
		}
	}()
	go func() {
		var err error
		defer func() {
			httpdone <- err
			close(httpdone)
		}()

		serveErrs := httpserver.Serve()

		for err = range serveErrs {
			if IsClose(err) {
				err = nil
			} else {
				httpErrorLogger.Err(err.Error())
			}
		}
	}()
	logger.Notice(fmt.Sprintf("serving http traffic on %s", httpserver.listener.Addr()))

	done := make(chan error, 1)
	defer func() {
		for err = range done {
			if err != nil {
				logger.Crit(err.Error())
				os.Exit(1)
			}
		}
	}()
	go func() {
		var err error
		defer func() {
			done <- err
			close(done)
		}()

		for sig := range sigch {
			logger.Notice(fmt.Sprintf("received signal: %s", sig))

			err = httpserver.Close()
			if err != nil {
				logger.Crit(fmt.Sprintf("error shutting down http server: %s", err))
				done <- err
			}

			err = statdaemon.Stop()
			if err != nil {
				// this is not a critical error
				logger.Notice(fmt.Sprintf("error shutting down stat daemon: %s", err))
				err = nil
			}

			signal.Stop(sigch)
			break
		}
	}()
}

func Routes() http.Handler {
	accesslogger, err := log.NewSyslog(syslog.LOG_NOTICE, "access")
	if err != nil {
		panic(err)
	}

	rBlockPath := regexp.MustCompile("(/[a-zA-Z0-9-_]+)+")

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		method, path := req.Method, req.URL.Path
		accesslogger.Notice(fmt.Sprintf("%s %q", method, path))
		switch {
		case method == "POST" && path == "/":
			CreateBlock(resp, req)
		case method == "DELETE" && rBlockPath.MatchString(path):
			DeleteBlock(resp, req)
		case method == "GET" && rBlockPath.MatchString(path):
			GetBlock(resp, req)
		default:
			http.NotFound(resp, req)
		}
	})

	return http.DefaultServeMux
}

func GetBlock(resp http.ResponseWriter, req *http.Request) {
	blockid := req.URL.Path[1:]
	fmt.Fprint(resp, blockid)
}

func CreateBlock(resp http.ResponseWriter, req *http.Request) {
	err := authenticate(req)
	if err != nil {
		http.Error(resp, "forbidden", http.StatusForbidden)
		return
	}

	_length := req.Header.Get("Content-Length")
	if _length == "" {
		http.Error(resp, "missing header: Content-Length", http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(_length, 10, 64)
	if err != nil {
		http.Error(resp, "invalid header: Content-Length", http.StatusBadRequest)
	}

	digest := req.Header.Get("Digest")
	if digest == "" {
		http.Error(resp, "missing header: Digest", http.StatusBadRequest)
		return
	}
	digestPattern := regexp.MustCompile(`^SHA=(\S)$`)
	matches := digestPattern.FindStringSubmatch(digest)
	if len(matches) > 0 {
		digest = matches[0]
	} else {
		http.Error(resp, "bad digest", http.StatusBadRequest)
		return
	}

	// read request data and fork to file system and checksum
	defer req.Body.Close()
	hash := sha1.New()
	tee := newTWriter(hash, ioutil.Discard)
	err = copyN(tee, req.Body, length)
	if err == errTooLarge {
		http.Error(resp, "post body too large", http.StatusRequestEntityTooLarge)
	}
	if err == errUnexpectedEOF {
		http.Error(resp, "unexpected end of block", http.StatusBadRequest)
	}
	if err != nil {
		http.Error(resp, "internal failure", http.StatusInternalServerError)
		return
	}
	computedDigest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if digest != computedDigest {
		http.Error(resp, "digest mismatch", http.StatusBadRequest)
		return
	}

	fmt.Println(resp, length)
}

func DeleteBlock(resp http.ResponseWriter, req *http.Request) {
	err := authenticate(req)
	if err != nil {
		http.Error(resp, "forbidden", http.StatusForbidden)
		return
	}

	blockid := req.URL.Path[1:]
	fmt.Fprint(resp, blockid)
}

func authenticate(req *http.Request) error {
	return nil
}

var errTooLarge = fmt.Errorf("too large")
var errUnexpectedEOF = fmt.Errorf("unexpected eof")

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
			return errTooLarge
		}
	}

	if totalread < n {
		return errUnexpectedEOF
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
