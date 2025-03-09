package worker

import (
	"context"
	"log"

	pb "go-grpc-mvp/pb"
)

type Worker struct{}

func (w *Worker) ProcessRequest(ctx context.Context, req *pb.BuildRequest) (*pb.GetBuildStatus, error) {
	log.Printf("Worker processing request for: %s", req.GetName())
	return &pb.HelloReply{Message: "Processed by worker: " + req.GetName()}, nil
}
