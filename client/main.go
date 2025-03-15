package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	pb "remote-build/remote-build"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = flag.String("addr", "localhost:50051", "The server address")

func main() {
	flag.Parse()

	files := []string{"main.c", "main2.c", "main3.c", "main4.c", "main5.c"}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMicServiceClient(conn)

	for _, filename := range files {
		fileData, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("Skipping file %s: %v", filename, err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Send build request
		resp, err := client.StartBuild(ctx, &pb.BuildRequest{
			Filename:    filename,
			FileContent: fileData,
		})
		if err != nil {
			log.Printf("Error while starting build for %s: %v", filename, err)
			continue
		}

		outputFile := strings.TrimSuffix(resp.Filename, ".c") + ".o"
		err = os.WriteFile(outputFile, resp.CompiledContent, 0755)
		if err != nil {
			log.Printf("Failed to save compiled file: %v", err)
		} else {
			log.Printf("Build Completed: %s saved as %s", filename, outputFile)
		}
	}
}
