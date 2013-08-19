// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// http_api.go [created: Sun, 18 Aug 2013]

package main
import (
	"fmt"
	"io"
	"github.com/bmatsuo/go-syslog"
	"schnutil/log"
	"net/http"
	"regexp"
	"strconv"
)

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

	block, err := NewService().Get(blockid)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := block.Open()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	defer body.Close()

	_, err = io.Copy(resp, body)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
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

	_, err = NewService().Create(req.Body, length, digest)
	req.Body.Close()
	switch err {
	case nil:
		fmt.Println(resp, length)
	case ErrTooLarge:
		http.Error(resp, "post body too large", http.StatusRequestEntityTooLarge)
	case ErrUnexpectedEOF:
		http.Error(resp, "unexpected end of block", http.StatusBadRequest)
	case ErrDigestMismatch:
		http.Error(resp, err.Error(), http.StatusBadRequest)
	default:
		http.Error(resp, "internal failure", http.StatusInternalServerError)
	}
}

func DeleteBlock(resp http.ResponseWriter, req *http.Request) {
	err := authenticate(req)
	if err != nil {
		http.Error(resp, "forbidden", http.StatusForbidden)
		return
	}

	blockid := req.URL.Path[1:]

	err = NewService().Delete(blockid)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func authenticate(req *http.Request) error {
	return nil
}
