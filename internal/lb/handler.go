package lb

import (
	"context"
	"io"
	"log"
	"sync"

	pb "github.com/akash-biswas-dev/grpc-lb/gen/lb/v1"
)

type LoadBalancerServer struct {
	pb.UnimplementedLoadBalancerServiceServer
	// Registry of active backends (Service B)
	backends sync.Map
	// List of active client streams (Service A) to push updates to
	watchers sync.Map
}

func NewLoadBalancerServer() *LoadBalancerServer {
	return &LoadBalancerServer{}
}

func (s *LoadBalancerServer) GetServers(con context.Context, r *pb.GetServersRequest) (*pb.GetServersResponse, error) {

	availableServers := pb.GetServersResponse{
		Address: []string{
			"127.0.0.1",
		},
	}

	return &availableServers, nil
}

func (s *LoadBalancerServer) Register(stream pb.LoadBalancerService_RegisterServer) error {

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(nil)
		}

		if err != nil {
			return err
		}

		serviceId := req.ServiceId
		ipAddress := req.Address

		log.Printf("Added POD with service-id %s, Ip-Address %s  ", serviceId, ipAddress)
	}
}

func (s *LoadBalancerServer) WatchServer(req *pb.WatchServerRequest, stream pb.LoadBalancerService_WatchServerServer) error {
	return nil
}

// Register implementation for Service B
