package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/redis/go-redis/v9"
)

var (
	kafkaBroker = "localhost:9092"
	redisAddr   = "localhost:6379"
	redisClient *redis.Client
)

func main() {
	// Initialize Redis Client
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Initialize Kafka Consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaBroker,
		"group.id":           "worker-group",
		"auto.offset.reset":  "earliest",
		"fetch.wait.max.ms":  "10",
		"fetch.min.bytes":    "1",
		"enable.auto.commit": "false",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Subscribe to Kafka Topic
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

		// Compute hash-based key for caching
		buildHash := computeHash(msg.Value)
		cacheKey := "cache:" + buildHash

		ctx := context.Background()

		// Check if compiled file exists in Redis cache
		cachedContent, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			log.Printf("Cache hit: Sending compiled file from Redis for %s (hash: %s)", filename, buildHash)
			sendToKafka(producer, filename, cachedContent)
			continue
		}

		// Process file and compile
		_, compiledContent, err := processFileFromKafka(filename, msg.Value)
		elapsedTime := time.Since(startTime)

		if err != nil {
			log.Printf("Compilation failed for %s (took %v): %v", filename, elapsedTime, err)
			continue
		}

		log.Printf("Compiled %s successfully in %v", filename, elapsedTime)

		// Store only the compiled `.o` file in Redis (cache for 10 minutes)
		err = redisClient.Set(ctx, cacheKey, compiledContent, 10*time.Minute).Err()
		if err != nil {
			log.Printf("Failed to cache compiled file in Redis: %v", err)
		} else {
			log.Printf("Cached compiled file for %s (hash: %s)", filename, buildHash)
		}

		// Send compiled file to Kafka
		sendToKafka(producer, filename, compiledContent)

		// Commit Kafka message offset
		_, err = consumer.CommitMessage(msg)
		if err != nil {
			log.Printf("Failed to commit Kafka message offset for %s: %v", filename, err)
		} else {
			log.Printf("Committed Kafka message offset for %s", filename)
		}
	}
}

// Compute a SHA-256 hash for file content
func computeHash(sourceCode []byte) string {
	hasher := sha256.New()
	hasher.Write(sourceCode)
	hashValue := hex.EncodeToString(hasher.Sum(nil))
	return hashValue
}

// sendToKafka sends the compiled file back to Kafka
func sendToKafka(producer *kafka.Producer, filename string, compiledContent []byte) {
	deliveryChan := make(chan kafka.Event)

	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &[]string{"build_responses"}[0], Partition: kafka.PartitionAny},
		Key:            []byte(filename),
		Value:          compiledContent,
	}, deliveryChan)

	if err != nil {
		log.Printf("Failed to produce Kafka message for %s: %v", filename, err)
		return
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		log.Printf("Failed to deliver message for %s: %v", filename, m.TopicPartition.Error)
	} else {
		log.Printf("Successfully sent compiled file for %s to Kafka", filename)
	}
}

// processFileFromKafka compiles C code received from Kafka
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

	// Read compiled .o file content
	compiledContent, err := os.ReadFile(objectFile)
	if err != nil {
		log.Printf("Failed to read compiled file: %s", objectFile)
		return "", nil, err
	}

	return objectFile, compiledContent, nil
}
