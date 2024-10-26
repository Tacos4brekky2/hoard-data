package main

import (
	//"google.golang.org/grpc/credentials/insecure"
	"context"
	"fmt"
	//"bytes"
	//"encoding/csv"
	//encoding/json"
	//"io"
	"time"
	//    "log"
	"net/http"
	//    "google.golang.org/grpc"
	//"flag"
	pb "github.com/Tacos4brekky2/hoard-data/api/grpc"
    poly "github.com/Tacos4brekky2/hoard-data/api/polygon"
    polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net"
	"os"
)

type Server struct {
	pb.UnimplementedMarketDataServiceServer
	MongoClient   *mongo.Client
	RedisClient   *redis.Client
	RequestClient http.Client
	PolygonClient *polygon.Client
}

// TODO:
// Splits, dividends, events, market status and holidays, etf updating, asset types
func (s *Server) GetAssets(ctx context.Context, req *pb.AssetRequest) (*pb.AssetResponse, error) {
	return &pb.AssetResponse{}, nil
}

// Fetches ticker info if no records are present
// TODO: Set update schedules and check date entered against current date to see if an update is needed
func (s *Server) updateTickerInfo(ctx context.Context, symbol string) error {
	var result bson.M
	err := s.MongoClient.Database("stonksdev").Collection("tickers").FindOne(ctx, bson.M{"symbol": symbol, "type": "ticker_info"}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			params := &models.GetTickerDetailsParams{
				Ticker: symbol,
			}
			info, err := s.PolygonClient.GetTickerDetails(ctx, params)
			if err != nil {
				return err
			}
            resp, err := s.MongoClient.Database("stonksdev").Collection("tickers").InsertOne(ctx, bson.M{"symbol": symbol, "type": "ticker_info", "last_modified": time.Now().UTC(), "data": info.Results})
			if err != nil {
				fmt.Println(err)
				return err

			}
			fmt.Printf("Successfully inserted info for symbol: %s   |   %o", symbol, resp)
			return nil
		}
	}

	//resp, err := s.MongoClient.Database("stonksdev").Collection("tickers").ReplaceOne(ctx, bson.M{"symbol": symbol}, info)
	//if err != nil {
	//		fmt.Println(err)
	//	return err
	//}
	fmt.Printf("Ticker info up to date")
	return nil
}

// Fetches financial info if no records are present
// TODO: Set update schedules and check date entered against current date to see if an update is needed
func (s *Server) updateFinancials(ctx context.Context, symbol string) error {
	var result bson.M
	err := s.MongoClient.Database("stonksdev").Collection("tickers").FindOne(ctx, bson.M{"symbol": symbol, "type": "financial_statements"}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			financials, err := poly.GetFinancials(symbol)
			if err != nil {
				return err
			}
			resp, err := s.MongoClient.Database("stonksdev").Collection("tickers").InsertOne(ctx, financials)
			if err != nil {
				fmt.Println(err)
				return err

			}
			fmt.Printf("Successfully inserted financial statements for symbol: %s   |   %o", symbol, resp)
			return nil
		}
	}

	//resp, err := s.MongoClient.Database("stonksdev").Collection("tickers").ReplaceOne(ctx, bson.M{"symbol": symbol}, info)
	//if err != nil {
	//		fmt.Println(err)
	//	return err
	//}
	fmt.Printf("Financial info up to date")
	return nil
}

// Fetches ticker price data if no records are found
// TODO: Check against last record's timestamp and update if out of date
func (s *Server) updateOHLCV(
	ctx context.Context,
	symbol string,
) error {
	var result bson.M
	err := s.MongoClient.Database("stonksdev").Collection("prices").FindOne(ctx, bson.M{"symbol": symbol}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No price data found, fetching all")
			data, err := poly.GetDailyOHLCV(symbol, 0)
			if err != nil {
				return fmt.Errorf("Error fetching data: %w", err)
			}
			insertResult, err := s.MongoClient.Database("stonksdev").Collection("prices").InsertMany(ctx, data)
			if err != nil {
				return fmt.Errorf("failed to insert documents: %w", err)
			}
			fmt.Printf("Inserted %d data points\n", len(insertResult.InsertedIDs))
			return nil
		}
	}
	fmt.Println("Price data up to date")
	return nil
}

func listenOnAddress(address string) (net.Listener, error) {
	// Listen on specified port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Listening on address: %v\n", address)
	return lis, nil
}

func connectMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to verify the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	fmt.Print("Connected to MongoDB!\n")
	return client, err
}

func run(ctx context.Context, req *pb.AssetRequest) error {
	// Create and ping a MongoDB client
	//mongoUri := *flag.String("mongouri", "mongodb://localhost:27017", "URI for the MongoDB server")
	mongoClient, err := connectMongoDB(ctx, "mongodb://localhost:27017")
	if err != nil {
		return err
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			fmt.Printf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	// Create gRPC server
	//grpcServer := grpc.NewServer()
	//pb.RegisterMarketDataServiceServer(grpcServer, &Server{MongoClient: mongoClient})

	// Create a new HTTP client
	timeout := time.Duration(5 * time.Second)
	requestClient := http.Client{
		Timeout: timeout,
	}

	polygonClient := polygon.New(os.Getenv("POLYGON_API_KEY"))

	serv := &Server{MongoClient: mongoClient, RedisClient: redisClient, RequestClient: requestClient, PolygonClient: polygonClient}
	err = serv.updateFinancials(ctx, "GME")
	if err != nil {
		return err
	}



	return nil
}

func main() {
	symbols := []string{"GOOG"}
	ctx := context.Background()
	req := &pb.AssetRequest{
		Id:      "1",
		Symbols: symbols,
	}
	if err := run(ctx, req); err == nil {
		fmt.Fprintf(os.Stderr, "error: v%\n", err)
		os.Exit(1)
	}
}
