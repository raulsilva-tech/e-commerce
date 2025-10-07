package main

import (
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/raulsilva-tech/e-commerce/services/order/config"
)

func getEnv(key, def string) string {

	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {

	cfg := config.Config{
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

	// kafkaWriter := producer.NewProducer("")
	// repo := repository.NewOrderRepository(dbConn)
	// uc := usecase.NewOrderUseCase(repo, kafkaWriter)

	//grpc server
	// grpcService := grpc.NewOrderService(*uc)
	// go grpcService.StartGRPCServer(cfg.GRPCServerPort)

	//web server
	// handler, err := webserver.NewServer(cfg, *uc)
	// if err != nil {
	// 	log.Fatalf("failed to create server: %v", err)
	// }

	// srv := &http.Server{
	// 	Addr:    ":" + cfg.WebServerPort,
	// 	Handler: handler,
	// }

	// go func() {
	// 	log.Printf("Auth webserver running on :%s", cfg.WebServerPort)
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("listen: %v", err)
	// 	}
	// }()

	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, os.Interrupt)
	// <-quit
	// log.Println("Shutting down server...")

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err := srv.Shutdown(ctx); err != nil {
	// 	log.Fatalf("Server forced to shutdown: %v", err)
	// }

	// log.Println("Server exiting")

}
