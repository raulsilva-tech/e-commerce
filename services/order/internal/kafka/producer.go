package producer

import (
	"context"
	"fmt"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func NewProducer(broker string) *kafka.Writer {

	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "order_created",
		Balancer: &kafka.LeastBytes{},
	}

}

func PublishOrderCreated(ctx context.Context, writer *kafka.Writer, orderID int64) error {

	msg := kafka.Message{
		Key:   []byte(strconv.Itoa(int(orderID))),
		Value: []byte(fmt.Sprintf(`{"order_id": "%d"}`, orderID)),
	}

	return writer.WriteMessages(ctx, msg)
}
