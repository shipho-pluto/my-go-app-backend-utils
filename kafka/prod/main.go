package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	topic      = "my-topic"
	brokerAddr = "localhost:9092"
	partitions = 2
	producers  = 2
)

type (
	Worker struct {
		workerPool  chan InputData
		wg          *sync.WaitGroup
		producerCnt int
	}

	InputData struct {
		produserID int
		data       int
	}
)

func NewWorker(producerCnt int) *Worker {
	return &Worker{
		workerPool:  make(chan InputData),
		wg:          &sync.WaitGroup{},
		producerCnt: producerCnt,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	createTopic(partitions)

	wg := sync.WaitGroup{}
	for i := range partitions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumer(ctx, i)
		}()
	}

	worker := NewWorker(producers)
	worker.balance(ctx)

	wg.Wait()
}

func (w *Worker) balance(ctx context.Context) {
	for producerID := range w.producerCnt {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for i := 0; ; i++ {
				select {
				case w.workerPool <- InputData{produserID: producerID + 1, data: i}:
				case <-ctx.Done():
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}

	for consumerID := range partitions {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for input := range w.workerPool {
				producer(ctx, consumerID, input)
			}
		}()
	}

	w.wg.Wait()
}

func producer(ctx context.Context, id int, input InputData) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddr},
		Topic:   topic,
		//Balancer: &kafka.Hash{},
	})
	defer writer.Close()

	key := fmt.Sprintf("key-%d", input.produserID)
	value := fmt.Sprintf("Message %d", input.data)
	msg := kafka.Message{
		Key:       []byte(key),
		Value:     []byte(value),
		Partition: id,
	}
	err := writer.WriteMessages(ctx, msg)

	if err != nil {
		log.Printf("[PRODUCER ERROR] %v", err)
		return
	}
	log.Printf("[PRODUCER INFO] key:%s", msg.Key)
}

func consumer(ctx context.Context, partition int) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddr},
		Topic:     topic,
		Partition: partition,
	})
	defer reader.Close()

	log.Printf("[CONSUMER] Starting for partition %d", partition)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[CONSUMER %d ERROR] %v", partition, err)
			continue
		}

		fmt.Printf("CONSUMER-%d: Value:\"%s\" Offset:%d Partition:%d\n",
			reader.Config().Partition, string(msg.Value), msg.Offset, msg.Partition)
	}
}

func createTopic(partitions int) {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_ = conn.DeleteTopics(topic)
	time.Sleep(2 * time.Second)

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Topic created successfully")
	time.Sleep(5 * time.Second)
}
