package grpc

import (
	"context"
	"log"
	"net"
	"strconv"

	"github.com/raulsilva-tech/e-commerce/services/product/internal/usecase"
	pb "github.com/raulsilva-tech/e-commerce/services/product/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	ProductUseCase usecase.ProductUseCase
	// cfg         config.Config
}

func NewProductServer(uc usecase.ProductUseCase) *ProductServer {
	return &ProductServer{
		ProductUseCase: uc,
	}
}

func (s *ProductServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {

	id, err := s.ProductUseCase.CreateProduct(ctx, req.Name, req.Price)
	if err != nil {
		return nil, err
	}

	return &pb.CreateProductResponse{
		Id: strconv.Itoa(int(id)),
	}, nil
}

func (s *ProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {

	id, _ := strconv.Atoi(req.Id)

	p, err := s.ProductUseCase.GetByProductId(ctx, int64(id))
	if err != nil || p == nil {
		return nil, err
	}

	return &pb.GetProductResponse{
		Id:    req.Id,
		Name:  p.Name,
		Price: p.Price,
	}, nil
}

func (s *ProductServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {

	list, err := s.ProductUseCase.ListProducts(ctx, 10)
	if err != nil {
		return nil, err
	}

	var listProductResponse = &pb.ListProductsResponse{}

	for _, p := range list {

		gpr := pb.GetProductResponse{Id: strconv.Itoa(int(p.ID)), Name: p.Name, Price: p.Price}

		listProductResponse.Products = append(listProductResponse.Products, &gpr)

	}

	return listProductResponse, nil
}

func (s *ProductServer) StartGRPCServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterProductServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Printf("Product gRPC server running on :%s", port)
	return grpcServer.Serve(lis)
}
