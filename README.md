# :telephone: minigo
From Minitel in Golang with :heart:

## Overview
Minigo provides an SDK to build a [Minitel](https://en.wikipedia.org/wiki/Minitel) server over:
* TCP `nightly`
* WebSockets `unavailable`
* Serial `unavailable`

## Teletel Standard
The [Teletel Standard](http://543210.free.fr/TV/stum1b.pdf) is ported to the Go language. The basic set of commands
is located within [teletel.go](minigo/main/teletel.go), the constants are located in [consts.go](minigo/main/consts.go).

The Teletel Standard commands are coded within bytes, 7 are reserved to the data, and the last one to the parity. 
The parity byte is either `1` when the parity is `odd` and `0` when `even`.

## Connect your Minitel
You can expose a TCP server and then redirect the traffic to your Minitel, either with [Asterisk](https://en.wikipedia.org/wiki/Asterisk_(PBX))
or to the serial port with `socat`:
```bash
socat /dev/<tty>,b1200,raw,echo=0 TCP:<IP>:3615
```