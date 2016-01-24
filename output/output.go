package output

import (
	"time"

	"github.com/Shopify/sarama"
)

// A KafkaProducer encapsulates a connection to a Kafka cluster.
type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string
}

// Returns an initialized KafkaProducer.
func NewKafkaProducer(brokers []string, topic string, bufferTime, bufferBytes int) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal     // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy // Compress messages
	config.Producer.Flush.Bytes = bufferBytes
	config.Producer.Flush.Frequency = time.Duration(bufferTime * 1000000)

	p, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	k := &KafkaProducer{
		producer: p,
		topic:    topic,
	}

	return k, nil
}

func (k *KafkaProducer) Write(s string) {
	k.producer.Input() <- &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(s),
	}
}

func (k *KafkaProducer) Close() error {
	return k.producer.Close()
}
