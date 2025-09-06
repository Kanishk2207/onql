package engine

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
)

type Msg = nats.Msg

type NatsClient struct {
	conn *nats.Conn
	js   nats.JetStreamContext
	ctx  context.Context
}

// Connect initializes the NATS connection.
// If useJetStream is true, it also initializes a JetStream context.
func (n *NatsClient) Connect(url string, useJetStream bool, opts ...nats.Option) error {
	n.ctx = context.Background()

	// connect to NATS
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return err
	}
	n.conn = conn

	// optionally initialize JetStream
	if useJetStream {
		js, err := conn.JetStream()
		if err != nil {
			conn.Close()
			return err
		}
		n.js = js
	}
	return nil
}

// Close gracefully drains and closes the NATS connection.
func (n *NatsClient) Close() error {
	if n.conn == nil {
		return nil
	}
	if err := n.conn.Drain(); err != nil {
		return err
	}
	n.conn.Close()
	return nil
}

// Publish sends a plain NATS message to the given subject.
func (n *NatsClient) Publish(subject string, data []byte) error {
	if n.conn == nil || !n.conn.IsConnected() {
		return errors.New("nats: not connected")
	}
	return n.conn.Publish(subject, data)
}

// Request sends a request and waits up to timeout for a single reply.
func (n *NatsClient) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	if n.conn == nil || !n.conn.IsConnected() {
		return nil, errors.New("nats: not connected")
	}
	return n.conn.Request(subject, data, timeout)
}

// Subscribe sets up a simple subscription on subject with the given handler.
func (n *NatsClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if n.conn == nil || !n.conn.IsConnected() {
		return nil, errors.New("nats: not connected")
	}
	return n.conn.Subscribe(subject, handler)
}

// JetStreamPublish publishes a message to a JetStream-enabled subject.
func (n *NatsClient) JetStreamPublish(subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	if n.js == nil {
		return nil, errors.New("nats: JetStream not initialized")
	}
	return n.js.Publish(subject, data, opts...)
}

// JetStreamSubscribe sets up a push-based durable subscription with manual acks.
func (n *NatsClient) JetStreamSubscribe(
	subject, durableName string,
	handler nats.MsgHandler,
) (*nats.Subscription, error) {
	if n.js == nil {
		return nil, errors.New("nats: JetStream not initialized")
	}
	return n.js.Subscribe(
		subject,
		handler,
		nats.Durable(durableName),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
	)
}

// JetStreamPullSubscribe sets up a pull-based durable subscription.
func (n *NatsClient) JetStreamPullSubscribe(
	subject, durableName string,
) (*nats.Subscription, error) {
	if n.js == nil {
		return nil, errors.New("nats: JetStream not initialized")
	}
	return n.js.PullSubscribe(
		subject,
		durableName,
		nats.PullMaxWaiting(128),
	)
}
