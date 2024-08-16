package kafka

import (
	"context"

	"github.com/s0vunia/platform_common/pkg/kafka/consumer"
)

// Consumer интерфейс потребитея
type Consumer interface {
	Consume(ctx context.Context, topicName string, handler consumer.Handler) (err error)
	Close() error
}
