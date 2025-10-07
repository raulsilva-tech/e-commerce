package webserver

import (
	"github.com/go-redis/redis/v8"
	"github.com/raulsilva-tech/e-commerce/services/order/config"
	"github.com/raulsilva-tech/e-commerce/services/order/internal/usecase"
)

type Server struct {
	cfg         config.Config
	authUseCase usecase.OrderUseCase
	rdb         *redis.Client
}
