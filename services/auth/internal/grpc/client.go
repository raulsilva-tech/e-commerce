package grpc

import (
	"context"
	"log"

	"github.com/raulsilva-tech/e-commerce/services/auth/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func attachJWT(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func CallSignup() {

	conn, err := grpc.NewClient("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	resp, err := client.Signup(context.Background(), &pb.SignupRequest{
		Email:    "example@test.com",
		Password: "123456",
	})

	if err != nil {
		log.Fatalf("error calling signup: %v", err)
	}

	log.Printf("Signup response: %v", resp)

}
