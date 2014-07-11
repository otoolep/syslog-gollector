gollector
========

*gollector* is a Syslog Collector, written in [Go](http://golang.org/), which has support for streaming received messages to [Apache Kafka](https://kafka.apache.org/), version 0.8. The messages can be written to Kafka in parsed format, or written exactly as received.

The logs lines must be [RFC5424](http://tools.ietf.org/html/rfc5424) compliant, and in the following format:

    <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROC-ID MSGID MSG"

Consult the RFC to learn what each of these fields is. The TIMESTAMP field must be in [RFC3339](http://www.ietf.org/rfc/rfc3339.txt) format. MSGID must be '-' (nil). Lines not matching this format are dropped by the gollector.

Multi-line Support
------------
The gollector supports multi-line messages, so messages such as stack traces will be considered a single message.

Parsing Mode
------------
Parsing mode is enabled by default. In this mode, the Syslog header is parsed, and the fields become keys in a JSON structure. This JSON structure is then written to Kafka. If parsing mode is not enabled, the log line is written to Kafka as it was received.

For example, imagine the following message is received by the gollector:

    <134>1 2013-09-04T10:25:52.618085 ubuntu sshd 1999 - password accepted for user root

With parsing disabled, the line is written as is to Kafka. With parsing enabled, the following JSON object is written to Kafka:

    {
        "priority":134,
        "version":1,
        "timestamp":"2013-09-04T10:25:52.618085",
        "host":"ubuntu","app":
        "sshd",
        "pid":1999,
        "message":
        "password accepted for user root"
    }

This parsed form may be useful to downstream consumers.


Building
------------

    git clone git@github.com:otoolep/gollector.git
    cd gollector
    export GOPATH=$PWD
    go get -d github.com/otoolep/gollector
    go build github.com/otoolep/gollector

Running
------------

Execute

        gollector -h

for command-line options. Make sure your Kafka cluster is up and running first.

Dependencies
------------
The most significant dependencies are:

* The Kafka 0.8 client [sarama](https://github.com/Shopify/sarama).
* The unit-test framework [Package check](https://gopkg.in/check.v1).

Thanks to the creators of both.

TODO
------------
This code is still work-in-progress. Key work items remaining include:

* Basic stats via a HTTP API on the gollector.
* Output to statsd.
* Handle errors from sarama QueueMessage() call.
* The gollector needs to be GC-profiled.
* Clean shutdown, including the use of control channels.
* UDP support.
* Support arbritrary MSGIDs.
