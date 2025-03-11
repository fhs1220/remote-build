package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"log"
	"net"
	"os/exec"
	"strings"

	"google.golang.org/grpc"
	pb "remote-build/remote-build"
)

var port = flag.Int("port", 50052, "The worker port") 

type WorkerServer struct {
	pb.UnimplementedWorkerServiceServer
}

func (w *WorkerServer) ProcessWork(ctx context.Context, req *pb.WorkRequest) (*pb.WorkResponse, error) {
	log.Printf("Worker on port %d received file: %s", *port, req.Filename)

	// Save the `.c` file
	err := os.WriteFile(req.Filename, req.FileContent, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	// Convert `.c` to `.o`
	compiledFilename := strings.Replace(req.Filename, ".c", ".o", 1)
	log.Printf("Compiling %s to %s", req.Filename, compiledFilename)

	// Compile using `gcc`
	cmd := exec.Command("gcc", "-c", req.Filename, "-o", compiledFilename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Compilation failed: %v, Output: %s", err, output)
	}

	compiledContent, err := os.ReadFile(compiledFilename)
	if err != nil {
		log.Fatalf("Failed to read compiled file: %v", err)
	}
	os.Remove(compiledFilename)

	log.Printf("Compilation successful: %s", compiledFilename)
	return &pb.WorkResponse{
		Filename:        compiledFilename,
		CompiledContent: compiledContent,
	}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", *port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWorkerServiceServer(grpcServer, &WorkerServer{})

	log.Printf("Worker is running on port %d...", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
