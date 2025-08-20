package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	topic         = "test-topic"
	brokerAddress = "localhost:9092"
)

func createTopic() {
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("Error creating topic: %v", err)
	} else {
		log.Println("Topic created successfully")
	}
}

func main() {
	createTopic()

	ctx := context.Background()
	go produce(ctx)
	consume(ctx)
}

func produce(ctx context.Context) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
	})

	for i := 0; ; i++ {
		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("Key-%d", i)),
			Value: []byte(fmt.Sprintf("Message-%d", i)),
		}
		err := w.WriteMessages(ctx, msg)
		if err != nil {
			log.Printf("[ERROR] Producer error: %v", err)
			continue
		}

		log.Printf("[INFO] Produced: %s", string(msg.Value))
		time.Sleep(1 * time.Second)
	}
}

func consume(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		GroupID: "test-group",
	})

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("[ERROR] Consumer error: %v", err)
			continue
		}

		log.Printf("[INFO] Consumed: key=%s value=%s partition=%d offset=%d",
			string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
	}
}
