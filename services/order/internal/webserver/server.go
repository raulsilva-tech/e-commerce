package webserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/raulsilva-tech/e-commerce/services/order/config"
	"github.com/raulsilva-tech/e-commerce/services/order/internal/usecase"
)

type Server struct {
	cfg          config.Config
	orderUseCase usecase.OrderUseCase
	rdb          *redis.Client
}

func NewServer(cfg config.Config, uc usecase.OrderUseCase) (http.Handler, error) {

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis ping: %v", err)
	}

	s := &Server{
		cfg:          cfg,
		orderUseCase: uc,
		rdb:          rdb,
	}
	r := mux.NewRouter()
	r.HandleFunc("/health", s.healthHandler).Methods("GET")

	r.HandleFunc("/orders/", s.CreateOrder).Methods("POST")
	// r.HandleFunc("/orders/:id", s.GetOrderById).Methods("GET")
	// r.HandleFunc("/orders/:limit", s.GetOrders).Methods("GET")
	// r.HandleFunc("/orders/:id", s.UpdateOrder).Methods("PUT")
	// r.HandleFunc("/orders/:id", s.DeleteOrder).Methods("DELETE")

	fs := http.FileServer(http.Dir("./docs"))
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", fs))

	return r, nil
}

type OrderDTO struct {
	ID        int64   `json:"id,omitempty"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
}

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	var order OrderDTO

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid input data: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := s.orderUseCase.CreateOrder(r.Context(), order.ProductID, order.Quantity, order.Total)
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
