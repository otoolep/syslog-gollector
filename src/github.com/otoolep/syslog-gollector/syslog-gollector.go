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
var tcpIface string
var udpIface string
var kBrokers string
var kBatch int
var kTopic string
var kBufferTime int
var kBufferBytes int
var pEnabled bool
var cCapacity int

// Types
const (
	connTcpHost      = "localhost:514"
	connUdpHost      = "localhost:514"
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
	flag.StringVar(&tcpIface, "tcp", connTcpHost, "TCP bind interface")
	flag.StringVar(&udpIface, "udp", connUdpHost, "UDP interface")
	flag.StringVar(&kBrokers, "broker", kafkaBrokers, "comma-delimited kafka brokers")
	flag.StringVar(&kTopic, "topic", kafkaTopic, "kafka topic")
	flag.IntVar(&kBatch, "batch", kafkaBatch, "Kafka batch size")
	flag.IntVar(&kBufferTime, "maxbuff", kafkaBufferTime, "Kafka client buffer max time (ms)")
	flag.IntVar(&kBufferBytes, "maxbytes", kafkaBufferBytes, "Kafka client buffer max bytes")
	flag.BoolVar(&pEnabled, "parse", parseEnabled, "enable syslog header parsing")
	flag.IntVar(&cCapacity, "chancap", chanCapacity, "channel buffering capacity")
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
		fmt.Println("Failed to create Kafka producer", err.Error())
		os.Exit(1)
	}
	log.Info("connected to kafka at %s", kBrokers)

	// Start the servers
	tcpServer := input.NewTcpServer(tcpIface)
	err = tcpServer.Start(func() chan<- string {
		return rawChan
	})
	if err != nil {
		fmt.Println("Failed to start TCP server", err.Error())
		os.Exit(1)
	}
	log.Info("listening on %s for TCP connections", tcpIface)

	udpServer := input.NewUdpServer(udpIface)
	err = udpServer.Start(func() chan<- string {
		return rawChan
	})
	if err != nil {
		fmt.Println("Failed to start UDP server", err.Error())
		os.Exit(1)
	}
	log.Info("listening on %s for UDP packets", udpIface)

	// Spin forever
	select {}
}
