package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
    //"google.golang.org/grpc/credentials/insecure"
	"log"
    "flag"
	pb "app/utils/userpb"
    //"app/utils/misc"
	"net"
	"time"
)

type Server struct {
    pb.UnimplementedDatabaseServiceServer
    mongoClient *mongo.Client
}

func (s *Server) ReadData(ctx context.Context, msg *pb.UserRequest) (*pb.DatabaseResponse, error) {
	return &pb.DatabaseResponse{}, nil
}

func (s *Server) WriteData(ctx context.Context, doc *pb.Document) (*pb.DatabaseResponse, error) {
//    timestamp := time.Now()

	return &pb.DatabaseResponse{}, nil
}

func main() {
	// Listen on specified port
    listenAddress := *flag.String("listenAddress", ":50051", "Address to reach the server")
    lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
    fmt.Printf("\nListening on address: %v", listenAddress)
    
    // Create and ping a MongoDB client
    mongoUri := *flag.String("mongouri", "mongodb:27017", "URI for the MongoDB server")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    clientOptions := options.Client().ApplyURI(mongoUri)
    mongoClient, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    err = mongoClient.Ping(ctx, nil)
    if err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)
    }
    fmt.Printf("\nConnected to MongoDB: %v", mongoUri)

	// Create gRPC server
	grpcServer := grpc.NewServer()
    pb.RegisterDatabaseServiceServer(grpcServer, &Server{mongoClient: mongoClient})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server on address %v: %v", listenAddress, err)
	}
}
