package lb

import (
	"log"
	"sync"

	pb "github.com/akash-biswas-dev/grpc-lb/gen/lb"
)

type LoadBalancerServer struct {
	pb.LoadBalancerServiceServer
	// Registry of active backends (Service B)
	backends sync.Map
	// List of active client streams (Service A) to push updates to
	watchers sync.Map
}

func NewLoadBalancerServer() *LoadBalancerServer {
	return &LoadBalancerServer{}
}

// WatchBackends implementation
func (s *LoadBalancerServer) WatchBackends(stream pb.LoadBalancerService_WatchServerServer) error {
	for {
		// Receive initial request from Service A
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		log.Printf("Client %s is watching %s", req.ClientId, req.ServiceToWatch)

		// TODO: Store this stream in s.watchers so when K8s finds a new Pod,
		// you can loop through these streams and call stream.Send(update)
	}
}

// Register implementation for Service B
func (s *LoadBalancerServer) Register(stream pb.LoadBalancerService_RegisterServer) error {
	// Handle Service B heartbeats here...
	return nil
}
