package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "remote-build/remote-build"

	"google.golang.org/grpc"
)

var port = flag.Int("port", 50052, "The worker port")

type WorkerServer struct {
	pb.UnimplementedWorkerServiceServer
}

// ProcessBuild handles the build request
func (w *WorkerServer) ProcessBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	buildID := fmt.Sprintf("%d", req.Name)

	log.Printf("Worker started processing build: %s (Build ID: %s)", req.Name, buildID)
	log.Printf("Worker finished processing build: %s (Build ID: %s)", req.Name, buildID)

	return &pb.BuildResponse{
		BuildId: buildID,
		Message: "Build Completed by Worker",
	}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWorkerServiceServer(grpcServer, &WorkerServer{})

	log.Printf("Worker is running on port %d...", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
