package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	pb "remote-build/remote-build"

	"google.golang.org/grpc"
)

var port = flag.Int("port", 50052, "The worker port")

type WorkerServer struct {
	pb.UnimplementedWorkerServiceServer
}

func (w *WorkerServer) ProcessWork(ctx context.Context, req *pb.WorkRequest) (*pb.WorkResponse, error) {
	log.Printf("Worker received files: %s", req.Files)
	files := strings.Split(req.Files, " ")
	for i, file := range files {
		if strings.HasSuffix(file, ".c") {
			files[i] = strings.Replace(file, ".c", ".o", 1)
		}
	}
	processedFiles := strings.Join(files, " ")

	log.Printf("Worker processed files: %s", processedFiles)
	return &pb.WorkResponse{ProcessedFiles: processedFiles}, nil
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
