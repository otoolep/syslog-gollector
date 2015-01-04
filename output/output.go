package output

import (
	log "code.google.com/p/log4go"

	"github.com/otoolep/sarama"
)

// A KafkaProducer encapsulates a connection to a Kafka cluster.
type KafkaProducer struct {
}

// Returns an initialized KafkaProducer.
func NewKafkaProducer(msgChan <-chan string, brokers []string, topic string, bufferTime, bufferBytes int) (*KafkaProducer, error) {
	self := &KafkaProducer{}

	clientConfig := sarama.NewClientConfig()
	client, err := sarama.NewClient("gocollector", brokers, clientConfig)
	if err != nil {
		log.Error("failed to create kafka client", err)
		return nil, err
	}

	producerConfig := sarama.NewProducerConfig()
	producerConfig.Partitioner = sarama.NewRandomPartitioner()
	producerConfig.MaxBufferedBytes = uint32(bufferBytes)
	producerConfig.MaxBufferTime = uint32(bufferTime)
	producer, err := sarama.NewProducer(client, producerConfig)
	if err != nil {
		log.Error("failed to create kafka producer", err)
		return nil, err
	}

	go func() {
		for message := range msgChan {
			producer.QueueMessage(topic, nil, sarama.StringEncoder(message))
		}
	}()

	go func() {
		for producer_error := range producer.Errors() {
			if producer_error != nil {
				log.Error("Error from producer.QueueMessage: ", producer_error)
			}
		}
	}()

	log.Info("kafka producer created")
	return self, nil
}
