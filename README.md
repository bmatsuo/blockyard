[1]: http://www.raspberrypi.org "Raspberry Pi homepage"
[2]: # "Architecture"

Blockyard
=========

Blockyard is a lightweight distributed file system designed to run on
[Raspberry Pi][1].

Yard is a shell script for building binaries and deploying blockyard
components to [Raspberry Nodes][2].

Usage
=====

The yard script must be executed inside a clone of the blockyard
repository.

    git clone git@github.com:bmatsuo/blockyard
    cd blockyard

Invoke the yard command as

    ./yard COMMAND [OPTIONS] [ARGUMENTS]

Build
-----

Build a the blockyard binaries.

    ./yard build [OS] [ARCH]

Dist
----

Create a distribution tar.gz file.

    ./yard dist [OS] [ARCH]

Deploy
------

Deploy to a node via ssh.

    ./yard deploy HOSTNAME DIST_ID

Docs
----

    ./yard docs [ -http=LADDR | IMPORT TAG ]

###Examples

    ./yard docs -http=:6060
    ./yard docs schnutil/log NewSyslog

Author
------

Bryan Matsuo [bryan dot matsuo at gmail dot com]

Copyright & License
-------------------

Copyright (c) 2013, Bryan Matsuo.
All rights reserved.
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
