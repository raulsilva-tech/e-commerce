package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/raulsilva-tech/e-commerce/services/order/internal/entity"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *entity.Order) error {
	_, err := r.db.Exec("INSERT INTO orders (product_id, quantity, total) VALUES ($1, $2, $3)",
		order.ProductID, order.Quantity, order.Total)
	return err
}
