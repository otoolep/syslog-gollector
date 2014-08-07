syslog-gollector
========

*syslog-gollector* is a Syslog Collector, written in [Go](http://golang.org/), which has support for streaming received messages to [Apache Kafka](https://kafka.apache.org/), version 0.8. The messages can be written to Kafka in parsed format, or written exactly as received. Check out [this blog post](http://www.philipotoole.com/writing-a-syslog-collector-in-go) for some background on this project.

The logs lines must be [RFC5424](http://tools.ietf.org/html/rfc5424) compliant, and in the following format:

    <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROC-ID MSGID MSG"

Consult the RFC to learn what each of these fields is. The TIMESTAMP field must be in [RFC3339](http://www.ietf.org/rfc/rfc3339.txt) format. Lines not matching this format are dropped by the syslog-gollector.

Checking out the "Running" section for hints on how to suitably configure Syslog clients.

Multi-line Support
------------
The syslog-gollector supports multi-line messages, so messages such as stack traces will be considered a single message.

Parsing Mode
------------
Parsing mode is enabled by default. In this mode, the Syslog header is parsed, and the fields become keys in a JSON structure. This JSON structure is then written to Kafka. If parsing mode is not enabled, the log line is written to Kafka as it was received.

For example, imagine the following message is received by the syslog-gollector:

    <134>1 2013-09-04T10:25:52.618085 ubuntu sshd 1999 - password accepted for user root

With parsing disabled, the line is written as is to Kafka. With parsing enabled, the following JSON object is written to Kafka:

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

This parsed form may be useful to downstream consumers.


Building
------------

    git clone git@github.com:otoolep/syslog-gollector.git
    cd syslog-gollector
    export GOPATH=$PWD
    go get -d github.com/otoolep/syslog-gollector
    go get gopkg.in/check.v1
    go build github.com/otoolep/syslog-gollector

Tested on 64-bit Kubuntu 14.04.

Running
------------

Execute

        syslog-gollector -h

for command-line options.

Make sure your Kafka cluster is up and running first. Point your syslog clients at the syslog-gollector, ensuring the message format is what syslog-gollector expects. Both [rsyslog](http://www.rsyslog.com/) and [syslog-ng](http://www.balabit.com/network-security/syslog-ng) support templating, which make it easy to format messages correctly. For example, an rsyslog template looks like so:

    $template SyslogGollector,"<%pri%>%protocol-version% %timestamp:::date-rfc3339% %HOSTNAME% %app-name% %procid% - %msg%\n"

syslog-ng looks like so:

    template SyslogGollector { template("<${PRI}>1 ${ISODATE} ${HOST} ${PROGRAM} ${PID} - $MSG\n"); template_escape(no) };

Dependencies
------------
The most significant dependencies are:

* The Kafka 0.8 client [sarama](https://github.com/Shopify/sarama).
* The unit-test framework [Package check](https://gopkg.in/check.v1).

Thanks to the creators of both.

TODO
------------
This code is still work-in-progress, and issues are being tracked. Other key tasks that span multiple issues include:

* Throughput needs to be measured.
* Run the program through Go's race-detector.

Miscellaneous
------------
Nothing to do with [gollector/gollector](https://github.com/gollector/gollector).
