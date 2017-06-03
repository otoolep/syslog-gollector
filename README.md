syslog-gollector [![Circle CI](https://circleci.com/gh/otoolep/syslog-gollector/tree/master.svg?style=svg)](https://circleci.com/gh/otoolep/syslog-gollector/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/otoolep/syslog-gollector)](https://goreportcard.com/report/github.com/otoolep/syslog-gollector) 
========

*Detailed background on syslog-gollector can be found on [these blog posts](http://www.philipotoole.com/tag/syslog-gollector/).*

*syslog-gollector* is a [Syslog](https://en.wikipedia.org/wiki/Syslog) Collector (sometimes called a Syslog Server), written in [Go](http://golang.org/) (golang), which has support for writing received log messages to [Apache Kafka](https://kafka.apache.org/), version 0.8. Log messages can be written to Kafka in parsed format, or written exactly as received.

The logs lines must be [RFC5424](http://tools.ietf.org/html/rfc5424) compliant, and in the following format:

    <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROC-ID MSGID MSG"

Consult the RFC to learn what each of these fields is. The TIMESTAMP field must be in [RFC3339](http://www.ietf.org/rfc/rfc3339.txt) format. Lines not matching this format are dropped by the syslog-gollector.

Check out the "Running" section for hints on how to easily configure Syslog clients to emit log mesages in the right format.

Multi-line Support
------------
The syslog-gollector supports multi-line log messages, so messages such as stack traces will be considered a single log message.

Parsing Mode
------------
Parsing mode is enabled by default. In this mode, the Syslog header is parsed, and the fields become keys in a JSON structure. This JSON structure is then written to Kafka. If parsing mode is not enabled, the log line is written to Kafka as it was received.

For example, imagine the following log line is received by the syslog-gollector:

    <134>1 2013-09-04T10:25:52.618085 ubuntu sshd 1999 - password accepted for user root

With parsing disabled, the line is written as-is to Kafka. With parsing enabled, the following JSON object is instead written to Kafka:

```json
{
    "priority":134,
    "version":1,
    "timestamp":"2013-09-04T10:25:52.618085",
    "host":"ubuntu",
    "app":"sshd",
    "pid":1999,
    "msgid": "-",
    "message": "password accepted for user root"
}
```

This parsed form may be useful to downstream consumers.

Building
------------
Tested on 64-bit Kubuntu 14.04.

```bash
mkdir ~/syslog-gollector # Or a directory of your choice.
cd ~/syslog-gollector
export GOPATH=$PWD
go get github.com/otoolep/syslog-gollector
```

To run the tests execute:
```bash
go get gopkg.in/check.v1
go test github.com/otoolep/syslog-gollector/...
```

If you want to hack on the source then modify it and rebuild like so (or whatever your Go workflow is):

```bash
cd $GOPATH/github.com/otoolep/syslog-gollector
....hack, hack,....
go install
```

Running
------------
The binary will be located in the ```$GOPATH/bin``` directory. Execute

```bash
syslog-gollector -h
```

for command-line options.

Make sure your Kafka cluster is up and running first. Point your syslog clients at the syslog-gollector, ensuring the log message format is what syslog-gollector expects. Both [rsyslog](http://www.rsyslog.com/) and [syslog-ng](http://www.balabit.com/network-security/syslog-ng) support templating, which make it easy to format messages correctly. For example, an rsyslog template looks like so:

    $template SyslogGollector,"<%pri%>%protocol-version% %timestamp:::date-rfc3339% %HOSTNAME% %app-name% %procid% - %msg%"

syslog-ng looks like so:

    template SyslogGollector { template("<${PRI}>1 ${ISODATE} ${HOST} ${PROGRAM} ${PID} - $MSG"); template_escape(no) };

Admin Control
------------
The syslog-gollector exposes a number of HTTP endpoints, for general statistics and diagnostics. This Admin server runs on localhost:8080 by default.

    /statistics
    /diagnostics

Adding the query parameter `pretty` to the URL will produce pretty-printed output. For example:

```bash
curl 'localhost:8080/statistics?pretty'
```

TODO
------------
This code is still work-in-progress, and issues are being tracked. Other key tasks that span multiple issues include:

* Throughput needs to be measured.
* Run the program through Go's race-detector.
