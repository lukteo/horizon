package messaging

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// NatsService handles interactions with NATS
type NatsService struct {
	Connection *nats.Conn
}

// NewNatsService creates a new instance of NatsService
func NewNatsService(natsURL string) (*NatsService, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NatsService{
		Connection: conn,
	}, nil
}

// Close closes the NATS connection
func (n *NatsService) Close() {
	if n.Connection != nil {
		n.Connection.Close()
	}
}

// PublishRawLog publishes a raw log entry to NATS
func (n *NatsService) PublishRawLog(subject string, data []byte) error {
	return n.Connection.Publish(subject, data)
}

// SubscribeToRawLogs subscribes to raw log entries from NATS
func (n *NatsService) SubscribeToRawLogs(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return n.Connection.Subscribe(subject, handler)
}