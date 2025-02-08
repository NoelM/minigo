# :telephone: minigo

From Minitel in Golang with :heart:

## Overview

Minigo provides an SDK to build a [Minitel](https://en.wikipedia.org/wiki/Minitel) server over:

* :white_check_mark: WebSockets
* :white_check_mark: TCP
* :white_check_mark: Serial Hayes Modem

## Teletel Standard

The [Teletel Standard](http://543210.free.fr/TV/stum1b.pdf) is ported to the Go language. The basic set of commands
is located within [teletel.go](minigo/main/teletel.go), the constants are located in [consts.go](minigo/main/consts.go).
