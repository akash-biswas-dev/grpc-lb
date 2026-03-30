package lb

import (
	"context"
	"errors"
	"io"
	"log"

	pb "github.com/akash-biswas-dev/grpc-lb/gen/lb/v1"
	sc "github.com/akash-biswas-dev/grpc-lb/internal/service-tracker"
)

type LoadBalancerServer struct {
	pb.UnimplementedLoadBalancerServiceServer
	serviceTracker sc.ServiceTracker
}

func NewLoadBalancerServer(expiration int) *LoadBalancerServer {
	return &LoadBalancerServer{
		serviceTracker: *sc.NewServiceTracker(expiration),
	}
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

	clientId, updateCh := s.serviceTracker.AddClient(req.ServiceId)

	// 2. Ensure we cleanup when the gRPC call ends (client disconnects)
	defer s.serviceTracker.RemoveClient(clientId, req.ServiceId)

	log.Printf("Client %s (ID: %s) started watching %s", req.ClientId, clientId, req.ServiceId)

	// 3. The Bridge: Loop and Select
	for {
		select {
		// CASE A: The Client disconnected or the connection timed out
		case <-stream.Context().Done():
			log.Printf("Client %s disconnected", clientId)
			return stream.Context().Err()

		// CASE B: Data arrived in the Go channel from UpdateNode or GetNode
		case update := <-updateCh:
			// Map your internal UpdateAction to the Protobuf Enum
			var grpcAction pb.WatchServerResponse_Action
			if update.Action == sc.Add {
				grpcAction = pb.WatchServerResponse_ACTION_ADD
			} else {
				grpcAction = pb.WatchServerResponse_ACTION_REMOVE
			}

			// 4. Send the data over the gRPC stream
			err := stream.Send(&pb.WatchServerResponse{
				Action: grpcAction,
				Instance: &pb.BackendInstance{
					Address: update.Address,
				},
			})

			if err != nil {
				log.Printf("Failed to send update to %s: %v", clientId, err)
				return err
			}
		}
	}
}

// Register implementation for Service B
func (s *LoadBalancerServer) GetAddress(_ context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {

	address, valid := s.serviceTracker.GetNode(req.ServiceId)

	if valid {
		return nil, errors.New("No nodes available.")
	}

	return &pb.GetAddressResponse{
		Address: address,
	}, nil
}
