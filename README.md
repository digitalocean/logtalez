# logtalez [![Build Status](https://travis-ci.org/digitalocean/logtalez.svg?branch=master)](https://travis-ci.org/digitalocean/logtalez) [![Doc Status](https://godoc.org/github.com/digitalocean/logtalez?status.png)](https://godoc.org/github.com/digitalocean/logtalez)

![logtalez](https://i.imgur.com/DFZxsBi.png)

## Problem Statement
We want to tail logs from remote servers as conveniently as if they were local, in a safe and secure manner.

## Solution
logtalez - a library and command line client for subscribing to log streams from rsyslog using the [omczmq](https://github.com/rsyslog/rsyslog/tree/master/contrib/omczmq) output plugin..

* Create dynamic topics using rsyslog's [parsing](http://www.rsyslog.com/doc/messageparser.html) and [template](http://www.rsyslog.com/doc/v8-stable/configuration/templates.html) features.
* Subscribe to topics to receive the logs you want.
* Publisher side filtering keeps bandwidth usage low.
* Brokerless design keeps operation simple.
* Ephemeral streaming keeps things light weight.
* [CurveZMQ](http://curvezmq.org/) authentication and encryption keeps things secure.

## Installation
### Dependencies
#### [libsodium](https://github.com/jedisct1/libsodium)
Version: 1.0.11 (or newer)

Sodium is a "new, easy-to-use software library for encryption, decryption, signatures, password hashing and more".  ZeroMQ uses sodium for the basis of the CurveZMQ security protocol.

#### [ZeroMQ](http://zeromq.org/) 
Version: 4.2.0 (or newer)

ZeroMQ is an embeddable [ZMTP](http://rfc.zeromq.org/spec:23) protocol library.

#### [CZMQ](http://czmq.zeromq.org/)
Version: 4.0.1 (or newer)

CZMQ is a high-level C binding for ZeroMQ.  It provides an API for various services on top of ZeroMQ such as authentication, actors, service discovery, etc.

#### [GoCZMQ](http://https://github.com/zeromq/goczmq)
GoCZMQ is a Go interface to the CZMQ API.

#### [Rsyslog](http://www.rsyslog.com/)
Version: 8.9.0 or newer

Rsyslog is the "rocket fast system for log processing".
You will need to use the "--enable-omczmq" configure flag to build zeromq + curve support.

### Generating Certificates
logtalez uses CURVE security certificates generated by the [zcert](http://api.zeromq.org/czmq3-0:zcert) API.  They are stored in [ZPL](http://rfc.zeromq.org/spec:4) format.  Logtalez includes a simple cert generation tool (curvecertgen) for convenience.

To generate a public / private key pair:

```
$ ./curvecertgen bogus_cert
Name: Brian
Email: bogus@whatever.com
Organization: Bogus Org
Version: 1
```

The above would generate a bogus_cert and bogus_cert_secret file.

### Configuring Your Rsyslog Server

The following rsyslog configuration snippet consists of:
* A template that dynamically sets a "topic" on a message consisting of hostname.syslogtag + an "@cee" cookie and JSON message payload
* A rule snippet that attempts to parse a syslog message as JSON, then outputs it over a zeromq publish socket using the template

```
module(load="mmjsonparse")
module(load="omczmq")

template(name="pubsub_host_tag" type="list") {
  property(name="hostname")
  constant(value=".")
  property(name="syslogtag" position.from="1" position.to="32")
  constant(value="@cee:")
  constant(value="{")
  constant(value="\"@timestamp\":\"")
  property(name="timereported" dateFormat="rfc3339" format="json")
  constant(value="\",\"host\":\"")
  property(name="hostname")
  constant(value="\",\"severity\":\"")
  property(name="syslogseverity-text")
  constant(value="\",\"facility\":\"")
  property(name="syslogfacility-text")
  constant(value="\",\"syslogtag\":\"")
  property(name="syslogtag" format="json")
  constant(value="\",")
  property(name="$!all-json" position.from="2")
} 

ruleset(name="zmq_pubsub_out") {
  action(
    name="zmq_pubsub"
    template="pubsub_host_tag"
    type="omczmq"
    endpoints="tcp://*:24444"
    socktype="PUB"
    authtype="CURVESERVER"
    clientcertpath="/etc/curve.d/"
    servercertpath="/etc/curve.d/my_server_cert"
  )
}

action(type="mmjsonprase")
if $parsesuccess == "OK" then {
  call zmq_pubsub_out
} 
```

## Usage

`````go
	import "github.com/digitalocean/logtalez"

	func main() {
		endpoints := []string{"tcp://127.0.0.1:24444,tcp://example.com:24444"}
		topics = []string{"host1.nginx","host2.nginx","host3.nginx"}

		serverCert := "/home/me/.curve/server_public_cert"
		clientCert := "/home/me/.curve/client_public_cert"

		lt, err := logtalez.New(endpoints, topics, serverCert, clientCert)
		if err != nil {
			panic(err)
		}

		buf := make([]byte, 65560)

		for {
			n, err := lt.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			fmt.Println(string(buf[:n]))
		}
	}
`````

## Tools That Work Well with Logtalez
* [jq](https://stedolan.github.io/jq/) JSON processor
* [humanlog](https://github.com/aybabtme/humanlog) "Logs for humans to read."
* Anything that can read stdout!

## GoDoc

[godoc](https://godoc.org/github.com/digitalocean/logtalez)

## License

This project uses the MPL v2 license, see LICENSE
