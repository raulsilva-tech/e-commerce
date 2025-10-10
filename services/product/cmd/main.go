package main

import (
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/raulsilva-tech/e-commerce/services/product/internal/grpc"
	"github.com/raulsilva-tech/e-commerce/services/product/internal/repository"
	"github.com/raulsilva-tech/e-commerce/services/product/internal/usecase"
)

type Config struct {
	WebServerPort   string
	DatabaseDSN     string
	RedisAddr       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	GRPCServerPort  string
}

func getEnv(key, def string) string {

	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {

	cfg := Config{
		WebServerPort:   getEnv("WEBSERVER_PORT", "8080"),
		DatabaseDSN:     getEnv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5433/ecommerce?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-prod"),
		GRPCServerPort:  getEnv("GRPCSERVER_PORT", "50051"),
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	dbConn, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer dbConn.Close()

	cache := repository.NewProductCache(cfg.RedisAddr)
	repo := repository.NewProductRepository(dbConn)
	uc := usecase.NewProductUseCase(repo, cache)

	//grpc server
	grpcService := grpc.NewProductServer(*uc)
	grpcService.StartGRPCServer(cfg.GRPCServerPort)

}
