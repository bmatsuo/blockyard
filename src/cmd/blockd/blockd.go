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
	"fmt"
	"github.com/bmatsuo/go-syslog"
	"os"
	"os/signal"
	"schnutil/log"
	"schnutil/stat"
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
			panic(e)
		} else {
			logger.Notice("shut down complete")
		}
		panic("shut down")
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
