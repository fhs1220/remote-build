package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	pb "remote-build/remote-build"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

// var (
// 	port       = flag.Int("port", 50051, "The server port")
// 	workerAddrs = flag.String("worker_addrs", "localhost:50052,localhost:50053,localhost:50054", "separated worker addresses")
// )

type Config struct {
	Config struct {
		NumberOfWorker int      `yaml:"numberofworker"`
		WorkerIP       []string `yaml:"workerIP"`
	} `yaml:"Config"`
}

type Server struct {
	pb.UnimplementedMicServiceServer
	workers    []pb.WorkerServiceClient
	mu         sync.Mutex
	nextWorker int
}

// Selects the next worker (RR)
func (s *Server) getNextWorker() pb.WorkerServiceClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	worker := s.workers[s.nextWorker]
	s.nextWorker = (s.nextWorker + 1) % len(s.workers)
	return worker
}

func (s *Server) StartBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	log.Printf("Server received file: %s", req.Filename)

	worker := s.getNextWorker()

	// resp, err := worker.ProcessWork(ctx, &pb.WorkRequest{
	// 	Filename:    req.Filename,
	// 	FileContent: req.FileContent,
	// })

	respChan := make(chan *pb.BuildResponse)
	errChan := make(chan error)

	go func() {
		resp, err := worker.ProcessWork(ctx, &pb.WorkRequest{
			Filename:    req.Filename,
			FileContent: req.FileContent,
		})

		if err != nil {
			errChan <- err
			return
		}
		respChan <- &pb.BuildResponse{
			Filename:        resp.Filename,
			CompiledContent: resp.CompiledContent,
		}
	}()

	// 	log.Printf("Server received compiled file: %s", resp.Filename)

	// 	return &pb.BuildResponse{
	// 		Filename:        resp.Filename,
	// 		CompiledContent: resp.CompiledContent,
	// 	}, nil
	// }

	select {
	case resp := <-respChan:
		log.Printf("Server received compiled file: %s", resp.Filename)
		return resp, nil
	case err := <-errChan:
		log.Printf("Error while processing file %s: %v", req.Filename, err)
		return nil, err
	case <-ctx.Done():
		log.Printf("Request for %s timed out", req.Filename)
		return nil, ctx.Err()
	}
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	flag.Parse()

	// workerAddrList := strings.Split(*workerAddrs, ",")
	// var workers []pb.WorkerServiceClient
	config, err := LoadConfig("server/config.yaml")
	if err != nil {
		log.Fatalf(" Failed to load config: %v", err)
	}
	log.Printf(" Loaded Config: %+v", config)
	var workers []pb.WorkerServiceClient

	// for _, addr := range config.Config.WorkerIP {
	// 	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// 	if err != nil {
	// 		log.Fatalf("Could not connect to worker %s: %v", addr, err)
	// 	}
	// 	workers = append(workers, pb.NewWorkerServiceClient(conn))
	// 	log.Printf("Connected to worker at %s", addr)
	// }

	for _, addr := range config.Config.WorkerIP {
		log.Printf("Attempting to connect to worker at %s...", addr)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf(" Could not connect to worker %s: %v", addr, err)
			continue 
		}
		workers = append(workers, pb.NewWorkerServiceClient(conn))
		log.Printf(" Successfully connected to worker at %s", addr)
	}

	if len(workers) == 0 {
		log.Fatalf("No workers connected!")
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{workers: workers}

	pb.RegisterMicServiceServer(grpcServer, server)

	log.Printf("Server is running on port 50051, distributing tasks to workers: %v", config.Config.WorkerIP)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
