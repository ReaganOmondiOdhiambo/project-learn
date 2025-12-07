package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

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
	// Start a custom span for the logic
	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "InternalLogic_CreateUser")
	defer span.End()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate work
	time.Sleep(50 * time.Millisecond)

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

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	}
	
	// Clean up http/https prefix for gRPC exporter which expects host:port
	// This is a naive cleanup, but works for the demo environment where we set http://jaeger:4317
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		endpoint = endpoint[7:]
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("go-server"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Add OpenTelemetry StatsHandler
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	
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
