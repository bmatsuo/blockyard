// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// stat.go [created: Fri, 16 Aug 2013]

// Package stat does ....
package stat

import (
	"fmt"
	"github.com/bmatsuo/go-syslog"
	"runtime"
	"schnutil/log"
	"time"
)

var ErrNotStarted = fmt.Errorf("not started")
var ErrStarted = fmt.Errorf("already started")

type RuntimeStatDaemon struct {
	running bool
	delay   time.Duration
	stop    chan chan chan error
}

func NewRuntimeStatDaemon(delay time.Duration) *RuntimeStatDaemon {
	d := new(RuntimeStatDaemon)
	d.delay = delay
	if d.delay == 0 {
		d.delay = 1 * time.Minute
	}

	d.stop = make(chan chan chan error, 1)
	d.stop <- nil

	return d
}

func (d *RuntimeStatDaemon) Start() error {
	logger, err := log.NewSyslog(syslog.LOG_NOTICE, "stats")
	if err != nil {
		return err
	}

	logger.Notice("start")

	stopch := <-d.stop

	if stopch != nil {
		return ErrStarted
	}

	stopch = make(chan chan error, 0)
	d.stop <- stopch
	var timer <-chan time.Time

	go func() {
		timerPrimer := make(chan time.Time, 1)
		timerPrimer <- time.Time{}
		timer = timerPrimer
		var lastSampleTime time.Time
		var lastTotalDur time.Duration
		var lastNumGC uint32
		for cont := true; cont; {
			select {
			case errch := <-stopch:
				errch <- nil
				cont = false
			case t := <-timer:
				timer = time.After(d.delay)
				now := time.Now()
				elapsedSec := float64(now.Sub(lastSampleTime)) / float64(time.Second)
				memstats := new(runtime.MemStats)
				runtime.ReadMemStats(memstats)

				if !t.IsZero() {
					numgo := runtime.NumGoroutine()
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.goroutines", numgo))

					deltaNumGC := memstats.NumGC - lastNumGC
					gcPerSecond := float64(deltaNumGC) / float64(elapsedSec)
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.gc.runs_per_second", gcPerSecond))

					gcDur := time.Duration(memstats.PauseTotalNs)
					deltaGcDur := gcDur - lastTotalDur
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.gc.total_time", deltaGcDur))
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.gc.total_time_per_second",
						deltaGcDur / time.Duration(elapsedSec)))

					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.allocated", memstats.Alloc))
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.mallocs", memstats.Mallocs))
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.frees", memstats.Frees))
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.heap_alloc", memstats.HeapAlloc))
					logger.Notice(fmt.Sprintf("%s %s %v",
						now, "runtime.stack_in_use", memstats.StackInuse))
				}

				lastSampleTime = now
				lastTotalDur = time.Duration(memstats.PauseTotalNs)
				lastNumGC = memstats.NumGC
			}
		}
	}()

	return nil
}

func (d *RuntimeStatDaemon) Stop() error {
	logger, err := log.NewSyslog(syslog.LOG_NOTICE, "stats.runtime")
	if err != nil {
		return err
	}

	logger.Notice("stop")

	stopch := <-d.stop
	defer func() { d.stop <- stopch }()

	if stopch == nil {
		return ErrNotStarted
	}

	errch := make(chan error, 1)
	stopch <- errch
	return <-errch
}
