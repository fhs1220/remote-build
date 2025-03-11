package main

import (
	"context"
	"flag"
	"log"
	"net"
	"sync"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "remote-build/remote-build"
)

var (
	port       = flag.Int("port", 50051, "The server port")
	workerAddrs = flag.String("worker_addrs", "localhost:50052,localhost:50053,localhost:50054", "Comma-separated worker addresses")
)

type Server struct {
	pb.UnimplementedMicServiceServer
	workers    []pb.WorkerServiceClient
	mu         sync.Mutex
	nextWorker int
}

// Selects the next worker (RR)
func (s *Server) getNextWorker() pb.WorkerServiceClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	worker := s.workers[s.nextWorker]
	s.nextWorker = (s.nextWorker + 1) % len(s.workers) 
	return worker
}

func (s *Server) StartBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	log.Printf("Server received file: %s", req.Filename)

	worker := s.getNextWorker()

	resp, err := worker.ProcessWork(ctx, &pb.WorkRequest{
		Filename:    req.Filename,
		FileContent: req.FileContent,
	})
	if err != nil {
		log.Printf("Error processing file %s: %v", req.Filename, err)
		return nil, err
	}

	log.Printf("Server received compiled file: %s", resp.Filename)

	return &pb.BuildResponse{
		Filename:        resp.Filename,
		CompiledContent: resp.CompiledContent,
	}, nil
}

func main() {
	flag.Parse()

	workerAddrList := strings.Split(*workerAddrs, ",")
	var workers []pb.WorkerServiceClient

	for _, addr := range workerAddrList {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not connect to worker %s: %v", addr, err)
		}
		workers = append(workers, pb.NewWorkerServiceClient(conn))
		log.Printf("Connected to worker at %s", addr)
	}

	if len(workers) == 0 {
		log.Fatalf("No workers connected!")
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{workers: workers}

	pb.RegisterMicServiceServer(grpcServer, server)

	log.Printf("Server is running on port %d, distributing tasks to workers: %v", *port, workerAddrList)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

