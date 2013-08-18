// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// http.go [created: Sat, 17 Aug 2013]

package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
)

type HTTPServer struct {
	Handler    http.Handler
	listener   net.Listener
	waitactive *sync.WaitGroup
}

func IsClose(err error) bool {
	if operr, ok := err.(*net.OpError); ok {
		errmsg := operr.Err.Error()
		if -1 < strings.Index(errmsg, "use of closed network connection") {
			err = nil
			return true
		}
	}
	return false
}

func NewHTTPServerAddr(laddr string) (*HTTPServer, error) {
	listnr, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return NewHTTPServer(listnr), nil
}

func NewHTTPServer(listnr net.Listener) *HTTPServer {
	server := new(HTTPServer)
	server.listener = listnr
	server.waitactive = new(sync.WaitGroup)
	return server
}

// an error received through the returned channel signals that the server
// entered into an unrecoverable state and halted.
func (server *HTTPServer) Serve() <-chan error {
	errch := make(chan error, 1)
	go func() {
		defer close(errch)

		_server := new(http.Server)

		_server.Handler = http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			server.waitactive.Add(1)
			defer server.waitactive.Done()

			server.Handler.ServeHTTP(resp, req)
		})

		err := _server.Serve(server.listener)
		if err != nil {
			errch <- err
		}
	}()
	return errch
}

func (server *HTTPServer) Close() error {
	err := server.listener.Close() // *net.Listener is threadsafe
	server.waitactive.Wait()
	return err
}
