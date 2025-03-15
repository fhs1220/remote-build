package main

import (
	"context"
	"flag"
	"log"
	"net"
	"sync"
	"time"

	pb "remote-build/remote-build"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/grpc"
)

var (
	addr        = flag.String("addr", ":50051", "The server address")
	kafkaBroker = flag.String("kafka-broker", "localhost:9092", "Kafka broker address")
)

type Server struct {
	pb.UnimplementedMicServiceServer
	producer      *kafka.Producer
	consumer      *kafka.Consumer
	responseCache map[string]*pb.BuildResponse
	cacheLock     sync.Mutex
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

	server := &Server{
		producer:      producer,
		consumer:      consumer,
		responseCache: make(map[string]*pb.BuildResponse),
	}

	go server.consumeBuildResponses()
	return server, nil
}

func (s *Server) StartBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	log.Printf("Received build request for file: %s", req.Filename)

	err := s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &[]string{"build_requests"}[0], Partition: kafka.PartitionAny},
		Key:            []byte(req.Filename),
		Value:          req.FileContent,
	}, nil)

	s.producer.Flush(1000)

	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return nil, err
	}

	// Wait for the worker response 
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		s.cacheLock.Lock()
		resp, exists := s.responseCache[req.Filename]
		if exists {
			delete(s.responseCache, req.Filename)
			s.cacheLock.Unlock()
			log.Printf("Sending compiled file: %s", resp.Filename)
			return resp, nil
		}
		s.cacheLock.Unlock()
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Timeout: No response for %s", req.Filename)
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
		log.Printf("Received compiled file from worker: %s", filename)

		s.cacheLock.Lock()
		s.responseCache[filename] = &pb.BuildResponse{
			Filename:        filename,
			CompiledContent: msg.Value,
		}
		log.Printf("Stored compiled file in cache: %s", filename)
		s.cacheLock.Unlock()
	}
}

func main() {
	flag.Parse()

	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer server.producer.Close()
	defer server.consumer.Close()

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
