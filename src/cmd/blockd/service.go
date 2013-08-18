// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// service.go [created: Sun, 18 Aug 2013]

package main

import (
	"io"
)

type Block interface {
	Id() string
	Body() io.ReadCloser
}

type API interface {
	Get(id string) (Block, error)
	Create(r io.Reader) (Block, error)
	Delete(id string) error
}
