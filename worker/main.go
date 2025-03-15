package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var kafkaBroker = "localhost:9092"

func main() {
	
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		"group.id":          "worker-group",
		"auto.offset.reset": "earliest",
		"fetch.wait.max.ms": "10",   
		"fetch.min.bytes": "1",      
		"enable.auto.commit": "false", 
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Subscribe the topic
	err = consumer.SubscribeTopics([]string{"build_requests"}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	// Initialize Kafka Producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	log.Println("Worker is listening for tasks...")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			continue
		}

		filename := string(msg.Key)

		startTime := time.Now()

		objectFile, compiledContent, err := processFileFromKafka(filename, msg.Value)
		elapsedTime := time.Since(startTime)

		if err != nil {
			log.Printf("Compilation failed for %s (took %v): %v", filename, elapsedTime, err)
			continue
		}

		log.Printf("Compiled %s successfully in %v", filename, elapsedTime)

		// Send compiled .o file to build_responses
		deliveryChan := make(chan kafka.Event) 
		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &[]string{"build_responses"}[0], Partition: kafka.PartitionAny},
			Key:            []byte(objectFile),
			Value:          compiledContent,
		}, deliveryChan)
			
		if err != nil {
			log.Printf("Failed to produce Kafka message for %s: %v", objectFile, err)
			continue 
		}

		e := <-deliveryChan
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			log.Printf("Failed to deliver message for %s: %v", objectFile, m.TopicPartition.Error)
			continue 
		}

		_, err = consumer.CommitMessage(msg)
		if err != nil {
			log.Printf("Failed to commit Kafka message offset for %s: %v", filename, err)
		} else {
			log.Printf("Committed Kafka message offset for %s", filename)
		}
	}
}

func processFileFromKafka(filename string, sourceCode []byte) (string, []byte, error) {
	log.Printf("Compiling file: %s (directly from Kafka)", filename)

	objectFile := strings.TrimSuffix(filename, ".c") + ".o"

	// Run GCC
	cmd := exec.Command("gcc", "-x", "c", "-c", "-", "-o", objectFile) 
	cmd.Stdin = bytes.NewReader(sourceCode)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("GCC Compilation Failed: %s", string(output))
		return "", nil, err
	}

	log.Printf("Compiled %s successfully", filename)

	// read compiled .o file content
	compiledContent, err := os.ReadFile(objectFile)
	if err != nil {
		log.Printf("Failed to read compiled file: %s", objectFile)
		return "", nil, err
	}

	return objectFile, compiledContent, nil
}

