// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// log.go [created: Wed, 14 Aug 2013]

/*
Unified logging for schnoodle.
*/
package log

import (
	"log"
	"os"
	"time"

	//"code.google.com/p/goprotobuf/proto"
	"github.com/bmatsuo/go-syslog"
	"github.com/bmatsuo/go-syslog/rsyslog"
)

var SyslogBase = os.Args[0]
var SyslogFacility = syslog.LOG_LOCAL3

func NewSyslog(pri syslog.Priority, name string) (*syslog.Writer, error) {
	fullname := SyslogBase
	if name != "" {
		fullname += "." + name
	}
	facility := pri & ^syslog.Priority(0x03)
	if facility != 0 {
		pri = pri | SyslogFacility
	}
	return syslog.DialAppended("", "", pri, fullname, rsyslog.Append)
}

type LogEventSink interface {
	Send(*LogEvent)
}

type LogEventSource interface {
	Recv(*LogEvent)
}

type Logger struct {
	sinks  map[string]LogEventSink
	source string
	tags   []string
}

func NewLogger(source string, tags ...string) *Logger {
	return &Logger{
		sinks:  make(map[string]LogEventSink, 0),
		source: source,
		tags:   tags,
	}
}

func (logger *Logger) Sink(name string, sink LogEventSink) {
	if sink == nil {
		delete(logger.sinks, name)
	} else {
		logger.sinks[name] = sink
	}
}

func (logger *Logger) Log(message string) {
	timestamp := time.Now().Format(time.RFC3339Nano)
	for _, sink := range logger.sinks {
		sink.Send(&LogEvent{
			Source:  &logger.source,
			Timestamp: &timestamp,
			Tags:    logger.tags,
			Message: &message,
		})
	}
}

type stdSink struct {
	*log.Logger
}

func (sink *stdSink) Send(event *LogEvent) {
	sink.Printf("%v %v %v %s", *event.Timestamp, *event.Source, event.Tags, *event.Message)
}

func StandardSink(logger *log.Logger) LogEventSink {
	return &stdSink{logger}
}
