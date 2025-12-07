package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/reagan/grpc-learning-project/server/pb"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedUserServiceServer
	mu    sync.Mutex
	users map[string]*pb.UserResponse
}

// CreateUser implements user.UserService
func (s *server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.UserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simple ID generation
	id := fmt.Sprintf("%d", len(s.users)+1)

	user := &pb.UserResponse{
		Id:    id,
		Name:  in.GetName(),
		Email: in.GetEmail(),
		Age:   in.GetAge(),
	}

	s.users[id] = user
	log.Printf("Created user: %v", user)
	return user, nil
}

// GetUser implements user.UserService
func (s *server) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.UserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[in.GetId()]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", in.GetId())
	}
	
	log.Printf("Fetched user: %v", user)
	return user, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	
	// Initialize the server struct
	userService := &server{
		users: make(map[string]*pb.UserResponse),
	}
	
	pb.RegisterUserServiceServer(s, userService)
	
	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
