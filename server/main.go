package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "remote-build/remote-build"
)

var (
	port       = flag.Int("port", 50051, "The server port")
	workerAddr = flag.String("worker_addr", "localhost:50052", "The worker address")
)

type Server struct {
	pb.UnimplementedMicServiceServer
	workerClient pb.WorkerServiceClient
}

func (s *Server) StartBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	log.Printf("Server received build request: %s", req.Files)

	resp, err := s.workerClient.ProcessWork(ctx, &pb.WorkRequest{Files: req.Files})
	if err != nil {
		return nil, fmt.Errorf("failed to process work: %v", err)
	}

	log.Printf("Server received processed files from Worker: %s", resp.ProcessedFiles)

	return &pb.BuildResponse{
		ResultFiles: resp.ProcessedFiles,
	}, nil
}

func main() {
	flag.Parse()

	workerConn, err := grpc.NewClient(*workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to worker: %v", err)
	}
	defer workerConn.Close()

	workerClient := pb.NewWorkerServiceClient(workerConn)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{workerClient: workerClient}

	pb.RegisterMicServiceServer(grpcServer, server)

	log.Printf("Server is running on port %d...", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
