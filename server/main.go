package main

import (
	"context"
	"flag"
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
	log.Printf("Server received file: %s", req.Filename)

	// Forward request to worker
	resp, err := s.workerClient.ProcessWork(ctx, &pb.WorkRequest{
		Filename:    req.Filename,
		FileContent: req.FileContent,
	})
	if err != nil {
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

	// Connect to the worker
	workerConn, err := grpc.NewClient(*workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to worker: %v", err)
	}
	defer workerConn.Close()

	workerClient := pb.NewWorkerServiceClient(workerConn)

	// Start the server
	lis, err := net.Listen("tcp", ":50051")
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
