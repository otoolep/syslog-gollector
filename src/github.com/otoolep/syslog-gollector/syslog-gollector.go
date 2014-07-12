package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/otoolep/syslog-gollector/input"
	"github.com/otoolep/syslog-gollector/output"

	log "code.google.com/p/log4go"
)

// Program parameters
var iface string
var kBrokers string
var kBatch int
var kTopic string
var kBufferTime int
var kBufferBytes int
var pEnabled bool
var cCapacity int

// Types
const (
	connHost         = "localhost:514"
	connType         = "tcp"
	kafkaBatch       = 10
	kafkaBrokers     = "localhost:9092"
	kafkaTopic       = "logs"
	kafkaBufferTime  = 1000
	kafkaBufferBytes = 512 * 1024
	parseEnabled     = true
	chanCapacity     = 0
)

func init() {
	flag.StringVar(&iface, "i", connHost, "bind interface")
	flag.StringVar(&kBrokers, "k", kafkaBrokers, "comma-delimited kafka brokers")
	flag.StringVar(&kTopic, "t", kafkaTopic, "kafka topic")
	flag.IntVar(&kBatch, "b", kafkaBatch, "Kafka batch size")
	flag.IntVar(&kBufferTime, "a", kafkaBufferTime, "Kafka client buffer max time (ms)")
	flag.IntVar(&kBufferBytes, "e", kafkaBufferBytes, "Kafka client buffer max bytes")
	flag.BoolVar(&pEnabled, "p", parseEnabled, "enable syslog header parsing")
	flag.IntVar(&cCapacity, "c", chanCapacity, "channel buffering capacity")
}

func main() {
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		log.Error("unable to determine hostname -- aborting")
		os.Exit(1)
	}
	log.Info("syslog server starting on %s, PID %d", hostname, os.Getpid())
	log.Info("machine has %d cores", runtime.NumCPU())

	// Log config
	log.Info("kafka brokers: %s", kBrokers)
	log.Info("kafka topic: %s", kTopic)
	log.Info("kafka batch size: %d", kBatch)
	log.Info("kafka buffer time: %dms", kBufferTime)
	log.Info("kafka buffer bytes: %d", kBufferBytes)
	log.Info("parsing enabled: %t", pEnabled)
	log.Info("channel buffering capacity: %d", cCapacity)

	// Prep the channels
	rawChan := make(chan string, cCapacity)
	prodChan := make(chan string, cCapacity)

	if pEnabled {
		// Feed the input through the Parser stage
		parser := input.NewRfc5424Parser()
		prodChan, err = parser.StreamingParse(rawChan)
	} else {
		// Pass the input directly to the output
		prodChan = rawChan
	}

	// Connect to Kafka
	_, err = output.NewKafkaProducer(prodChan, strings.Split(kBrokers, ","), kTopic, kBufferTime, kBufferBytes)
	if err != nil {
		log.Error("unable to create Kafka Producer", err)
		os.Exit(1)
	}
	log.Info("connected to kafka at %s", kBrokers)

	// Start the server
	tcpServer := input.NewTcpServer(iface)
	err = tcpServer.Start(func() chan<- string {
		return rawChan
	})
	if err != nil {
		fmt.Println("Failed to start server", err.Error())
		os.Exit(1)
	}
	log.Info("listening on %s for connections", iface)

	// Spin forever
	select {}
}
