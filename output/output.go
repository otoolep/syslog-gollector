package output

import (
	"time"

	"github.com/Shopify/sarama"

	metrics "github.com/rcrowley/go-metrics"
)

// A KafkaProducer encapsulates a connection to a Kafka cluster.
type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string

	registry metrics.Registry
	msgTx    metrics.Counter
	bytesTx  metrics.Counter
}

// NewKafkaProducer returns an initialized KafkaProducer.
func NewKafkaProducer(brokers []string, topic string, bufferTime, bufferBytes, batchSz int) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal     // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy // Compress messages
	config.Producer.Flush.Bytes = bufferBytes
	config.Producer.Flush.Frequency = time.Duration(bufferTime * 1000000)
	config.Producer.Flush.Messages = batchSz

	p, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	k := &KafkaProducer{
		producer: p,
		topic:    topic,
		registry: metrics.NewRegistry(),
		msgTx:    metrics.NewCounter(),
		bytesTx:  metrics.NewCounter(),
	}

	k.registry.Register("messages.transmitted", k.msgTx)
	k.registry.Register("messages.bytes.transmitted", k.bytesTx)

	return k, nil
}

func (k *KafkaProducer) Write(s string) {
	k.producer.Input() <- &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(s),
	}
	k.msgTx.Inc(1)
	k.bytesTx.Inc(int64(len(s)))
}

// Statistics returns an object storing statistics, which supports JSON
// marshalling.
func (k *KafkaProducer) Statistics() (metrics.Registry, error) {
	return k.registry, nil
}

// Close closes the producer.
func (k *KafkaProducer) Close() error {
	return k.producer.Close()
}
