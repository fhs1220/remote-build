package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	pb "remote-build/remote-build"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr        = flag.String("addr", "localhost:50051", "The server address")
	redisAddr   = flag.String("redis-addr", "localhost:6379", "Redis address")
	redisClient *redis.Client
)

func main() {
	flag.Parse()

	files := []string{"main.c", "main2.c", "main3.c", "main4.c", "main5.c"}

	// Initialize Redis Client
	redisClient = redis.NewClient(&redis.Options{
		Addr: *redisAddr,
		DB:   0,
	})

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMicServiceClient(conn)

	for _, filename := range files {
		ctx := context.Background()

		// Read file content
		fileData, err := os.ReadFile(filename)
		if err != nil {
			log.Printf("Skipping file %s: %v", filename, err)
			continue
		}

		// Compute a unique hash-based cache key
		buildHash := computeHash(fileData)
		cacheKey := "cache:" + buildHash

		// Check Redis cache using the hash key
		cachedContent, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			log.Printf("Cache hit in client: %s (retrieved from Redis, hash: %s)", filename, buildHash)
			saveCompiledFile(filename, cachedContent)
			continue
		}

		grpcCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Send build request to server
		resp, err := client.StartBuild(grpcCtx, &pb.BuildRequest{
			Filename:    filename,
			FileContent: fileData,
		})
		if err != nil {
			log.Printf("Error while starting build for %s: %v", filename, err)
			continue
		}

		// Save compiled file locally
		saveCompiledFile(filename, resp.CompiledContent)

	}
}

// Compute a hash for file content
func computeHash(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}

func saveCompiledFile(filename string, compiledContent []byte) {
	outputFile := strings.TrimSuffix(filename, ".c") + ".o"
	err := os.WriteFile(outputFile, compiledContent, 0755)
	if err != nil {
		log.Printf("Failed to save compiled file: %v", err)
	} else {
		log.Printf("Build Completed: %s saved as %s", filename, outputFile)
	}
}
