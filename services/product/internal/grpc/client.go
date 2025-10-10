package grpc

import (
	"context"
	"log"
	"time"

	pb "github.com/raulsilva-tech/e-commerce/services/product/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetProductByID(id string) (*pb.GetProductResponse, error) {

	conn, err := grpc.NewClient("auth-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		log.Printf("Error calling gRPC: %v", err)
		return nil, err
	}

	return resp, nil
}
