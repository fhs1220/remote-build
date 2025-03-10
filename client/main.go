package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "remote-build/remote-build"
)

var addr = flag.String("addr", "localhost:50051", "The server address")

func main() {
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMicServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.StartBuild(ctx, &pb.BuildRequest{Files: "main.c lib.c"})
	if err != nil {
		log.Fatalf("Error while starting build: %v", err)
	}
	log.Printf("Build Completed: %s", resp.ResultFiles)
}
