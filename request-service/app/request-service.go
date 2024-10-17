package main

import (
    "context"
    "fmt"
    "net/http"
    "log"
    "google.golang.org/grpc"
    pb "app/utils/userpb"
)

type server struct{
    pb.UnimplementedRequestServiceServer
}

func (*server) FetchData(ctx context.Context, params *pb.UserRequest) (*pb.Response, error) {
    req, _ := http.NewRequest(params.RequestMethod, params.Url, nil)
    req.Header.Add(params.GetHeaderKey())
	return response, nil
}

func main() {
    
}
