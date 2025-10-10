package usecase

import (
	"context"

	"github.com/raulsilva-tech/e-commerce/services/order/internal/entity"
	producer "github.com/raulsilva-tech/e-commerce/services/order/internal/kafka"
	"github.com/raulsilva-tech/e-commerce/services/order/internal/repository"
	"github.com/segmentio/kafka-go"
)

type OrderUseCase struct {
	repo     *repository.OrderRepository
	producer *kafka.Writer
}

func NewOrderUseCase(r *repository.OrderRepository, p *kafka.Writer) *OrderUseCase {
	return &OrderUseCase{repo: r, producer: p}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, productID int64, quantity int, total float64) error {
	order, err := entity.NewOrder(0, productID, quantity, total)
	if err != nil {
		return err
	}

	if err := uc.repo.Create(order); err != nil {
		return err
	}

	if err := producer.PublishOrderCreated(ctx, uc.producer, order.ID); err != nil {
		return err
	}

	return nil
}
