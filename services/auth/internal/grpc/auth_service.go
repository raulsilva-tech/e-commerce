package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/raulsilva-tech/e-commerce/services/auth/config"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/usecase"
	"github.com/raulsilva-tech/e-commerce/services/auth/internal/webserver"
	"github.com/raulsilva-tech/e-commerce/services/auth/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	AuthUseCase usecase.AuthUseCase
	cfg         config.Config
}

func NewAuthService(uc usecase.AuthUseCase) *AuthService {
	return &AuthService{
		AuthUseCase: uc,
	}
}

func (s *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {

	user, err := s.AuthUseCase.Login(usecase.LoginInput{Email: in.Email, Password: in.Password})
	if err != nil {
		return nil, err
	}

	accessToken, err := webserver.MakeAccessToken(user.ID, user.Email, s.cfg.AccessTokenTTL, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := webserver.MakeRefreshToken()
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Signup(ctx context.Context, in *pb.SignupRequest) (*pb.SignupResponse, error) {

	id, err := s.AuthUseCase.Signup(usecase.SignupInput{Name: in.Name, Email: in.Email, Password: in.Password})
	if err != nil {
		return nil, err
	}

	return &pb.SignupResponse{
		UserId: strconv.Itoa(int(id)),
		Email:  in.Email,
	}, nil
}

func (s *AuthService) StartGRPCServer(port string) error {

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen %w", err)
	}

	grpcServer := grpc.NewServer(
	// grpc.UnaryInterceptor(UnaryServerInterceptorJWT),
	)

	pb.RegisterAuthServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Printf("Auth gRPC server running on :%s", port)
	return grpcServer.Serve(lis)
}

func jwtInterceptor(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}

	authHeader := md["authorization"]
	if len(authHeader) == 0 {
		return nil, errors.New("missing authorization token")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	// Aqui você valida o JWT (expiração, assinatura, claims, etc)
	if token == "" {
		return nil, errors.New("invalid token")
	}

	return ctx, nil
}

func UnaryServerInterceptorJWT(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	newCtx, err := jwtInterceptor(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}
