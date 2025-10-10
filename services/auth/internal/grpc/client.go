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

func CallSignup(ctx context.Context, token string) error {

	conn, err := grpc.NewClient("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return err
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	resp, err := client.Signup(attachJWT(ctx, token), &pb.SignupRequest{
		Email:    "example@test.com",
		Password: "123456",
	})

	if err != nil {
		log.Fatalf("error calling signup: %v", err)
		return err
	}

	log.Printf("Signup response: %v", resp)
	return nil
}
