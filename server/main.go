package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"net"
	"sync"
	"time"

	pb "remote-build/remote-build"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

var (
	addr        = flag.String("addr", ":50051", "The server address")
	kafkaBroker = flag.String("kafka-broker", "localhost:9092", "Kafka broker address")
	redisAddr   = flag.String("redis-addr", "localhost:6379", "Redis address")
)

type Server struct {
	pb.UnimplementedMicServiceServer
	producer  *kafka.Producer
	consumer  *kafka.Consumer
	redis     *redis.Client
	cacheLock sync.Mutex
}

func NewServer() (*Server, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": *kafkaBroker})
	if err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": *kafkaBroker,
		"group.id":          "server-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	err = consumer.Subscribe("build_responses", nil)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
		DB:   0,
	})

	server := &Server{
		producer: producer,
		consumer: consumer,
		redis:    redisClient,
	}

	go server.consumeBuildResponses()
	return server, nil
}

// Compute a hash for a file's content (CAS - Content Addressable Storage)
func computeHash(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *Server) StartBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	log.Printf("Received build request for file: %s", req.Filename)

	// Compute a unique hash-based key
	buildHash := computeHash(req.FileContent)
	cacheKey := "cache:" + buildHash

	// Check if compiled `.o` file is already in Redis
	cachedContent, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		log.Printf("Cache hit: Returning compiled file from Redis for %s (hash: %s)", req.Filename, buildHash)
		return &pb.BuildResponse{Filename: req.Filename, CompiledContent: cachedContent}, nil
	}

	// Publish build request to Kafka
	err = s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &[]string{"build_requests"}[0], Partition: kafka.PartitionAny},
		Key:            []byte(req.Filename),
		Value:          req.FileContent,
	}, nil)

	s.producer.Flush(1000)

	if err != nil {
		log.Printf("Failed to send build request to Kafka: %v", err)
		return nil, err
	}

	// Wait for compiled `.o` file to be stored in Redis
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		cachedContent, err := s.redis.Get(ctx, cacheKey).Bytes()
		if err == nil {
			log.Printf("Build result found in Redis for %s (hash: %s)", req.Filename, buildHash)
			return &pb.BuildResponse{Filename: req.Filename, CompiledContent: cachedContent}, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Timeout: No response received for %s", req.Filename)
	return nil, context.DeadlineExceeded
}

func (s *Server) consumeBuildResponses() {
	log.Println("Server listening for compiled files...")

	for {
		msg, err := s.consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error: %v", err)
			continue
		}
		filename := string(msg.Key)
		buildHash := computeHash(msg.Value)
		log.Printf("Received compiled file from worker: %s (hash: %s)", filename, buildHash)

	}
}

func main() {
	flag.Parse()

	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.producer.Close()
	defer server.consumer.Close()
	defer server.redis.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMicServiceServer(grpcServer, server)

	log.Println("Server is running on port 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
