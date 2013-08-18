// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// service.go [created: Sat, 17 Aug 2013]

// Package service does ....
package service

import (
	"errors"
)

var ErrStopped = errors.New("sevice was stopped")

// The service must close chans returned by Start() when Stop() is called.
type Service interface {
	Start(errch chan<- error) error
	Stop() error
}

type Stack struct {
	errch    chan<- error
	services []Service
}

func (stack Stack) Start(errch chan<- error) error {
	if stack.errch != nil { // race!
		return errors.New("already started")
	}
	stack.errch = errch

	for i, service := range stack.services {
		err := service.Start(errch)
		if err != nil {
			// close successfully started services
			for j := i - 1; j >= 0; j-- {
				stack.services[j].Stop() // ignore errors
			}
			return err
		}
	}
	return nil
}

func (stack Stack) Stop() error {
	var err error
	for i := len(stack.services) - 1; i >= 0; i-- {
		_err := stack.services[i].Stop()
		if _err != nil {
			if err == nil {
				err = _err
			}
		}
	}
	return err
}
