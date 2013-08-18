Blockyard
=========

Lightweight block storage

Usage
-----

Blockyard does not support installation through `go install` for deployment
easy and stability. Instead clone the repository and use the `blockyard` tool
to build executables.

    git clone git@github.com:bmatsuo/blockyard
    cd blockyard
    ./blockyard build

Docs
----

    ./blockyard docs [ -http=LADDR | IMPORT TAG ]

###Examples

    ./blockyard docs -http=:6060
    ./blockyard docs schuntil/log NewSyslog

Author
------

Bryan Matsuo [bryan dot matsuo at gmail dot com]

Copyright & License
-------------------

Copyright (c) 2013, Bryan Matsuo.
All rights reserved.
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
