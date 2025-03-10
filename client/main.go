package main

import (
	"context"
	"flag"
	"os"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "remote-build/remote-build"
)

var addr = flag.String("addr", "localhost:50051", "The server address")

func main() {
	flag.Parse()

	// Read the actual C file
	filename := "main.c"
	fileData, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Connect to the server
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMicServiceClient(conn)

	// Send build request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.StartBuild(ctx, &pb.BuildRequest{
		Filename:    filename,
		FileContent: fileData,
	})
	if err != nil {
		log.Fatalf("Error while starting build: %v", err)
	}

	// Save the compiled `.o` file
	outputFilename := resp.Filename
	err = os.WriteFile(outputFilename, resp.CompiledContent, 0644)
	if err != nil {
		log.Fatalf("Failed to save compiled file: %v", err)
	}

	log.Printf("Build Completed: %s saved as %s", filename, outputFilename)
}
