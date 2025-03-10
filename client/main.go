package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "remote-build/remote-build"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = flag.String("addr", "localhost:50051", "The address to connect to")

func main() {
	flag.Parse()

	// Connect to the server
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMicServiceClient(conn)

	// Send build request
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.StartBuild(ctx, &pb.BuildRequest{Name: "MyProject"})
	if err != nil {
		log.Fatalf("Error while starting build: %v", err)
	}
	log.Printf("Build Completed: %s (ID: %s)", resp.Message, resp.BuildId)
}
