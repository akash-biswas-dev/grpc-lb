package main

import (
	"log"
	"net"

	pb "github.com/akash-biswas-dev/grpc-lb/gen/lb/v1"
	"github.com/akash-biswas-dev/grpc-lb/internal/lb"
	"google.golang.org/grpc"
)

func main() {
	// 1. Listen on a TCP port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 2. Create the gRPC server
	grpcServer := grpc.NewServer()

	// 3. Initialize your LoadBalancer implementation
	// This struct will hold the logic for Register and WatchBackends
	lbService := lb.NewLoadBalancerServer(30)

	// 4. Register the service with the gRPC server
	pb.RegisterLoadBalancerServiceServer(grpcServer, lbService)

	log.Printf("NexusSphere LB started on %v", lis.Addr())

	// 5. Serve requests
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
