package iotaconn_grpc

import (
	"context"
	"io"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/iotaledger/hive.go/log"
)

const defaultBuf = 64

// StreamClient is a generic, reconnecting server-stream consumer.
type StreamClient[T any] struct {
	name    string // for logging
	address string
	options []grpc.DialOption
	logger  log.Logger

	conn *grpc.ClientConn

	// openStream must use conn to build the correct service client and open the server stream.
	// It returns a Recv function that yields messages of type T.
	openStream func(ctx context.Context, conn *grpc.ClientConn) (recv func() (T, error), err error)

	events chan T
	wg     sync.WaitGroup

	reconnectInterval time.Duration
	maxReconnectDelay time.Duration
}

func NewStreamClient[T any](
	name string,
	address string,
	logger log.Logger,
	openStream func(ctx context.Context, conn *grpc.ClientConn) (recv func() (T, error), err error),
	opts ...grpc.DialOption,
) *StreamClient[T] {
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			},
		),
	}
	all := append(defaultOpts, opts...)
	return &StreamClient[T]{
		name:              name,
		address:           address,
		options:           all,
		logger:            logger,
		events:            make(chan T, defaultBuf),
		reconnectInterval: 5 * time.Second,
		maxReconnectDelay: 30 * time.Second,
		openStream:        openStream,
	}
}

func (c *StreamClient[T]) Start(ctx context.Context) <-chan T {
	c.wg.Add(1)
	go c.run(ctx)
	return c.events
}

func (c *StreamClient[T]) WaitUntilStopped() {
	c.wg.Wait()
}

func (c *StreamClient[T]) run(ctx context.Context) {
	defer c.wg.Done()
	defer close(c.events)

	if err := c.dial(); err != nil {
		c.logger.LogErrorf("%s: gRPC dial failed: %v", c.name, err)
	}

	backoff := c.reconnectInterval

	for {
		if ctx.Err() != nil {
			break
		}

		if c.conn == nil {
			if !sleepCtx(ctx, backoff) {
				break
			}
			backoff = c.increaseBackoff(backoff)
			_ = c.dial()
			continue
		}

		recv, err := c.openStream(ctx, c.conn)
		if err != nil {
			c.logger.LogErrorf("%s: failed to create stream: %v", c.name, err)
			if !sleepCtx(ctx, backoff) {
				break
			}
			backoff = c.increaseBackoff(backoff)
			continue
		}

		c.logger.LogInfof("%s: stream established", c.name)
		backoff = c.reconnectInterval

		for {
			if ctx.Err() != nil {
				break
			}
			msg, err := recv()
			if err == io.EOF {
				c.logger.LogErrorf("%s: stream ended", c.name)
				break
			}
			if err != nil {
				c.logger.LogErrorf("%s: stream receive error: %v", c.name, err)
				break
			}

			select {
			case c.events <- msg:
			case <-ctx.Done():
				break
			}
		}

		if !sleepCtx(ctx, backoff) {
			break
		}

		backoff = c.increaseBackoff(backoff)
	}

	conn := c.conn
	c.conn = nil
	if conn != nil {
		_ = conn.Close()
	}
}

func (c *StreamClient[T]) dial() error {
	if c.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(c.address, c.options...)
	if err != nil {
		c.logger.LogErrorf("%s: failed to connect to %s: %v", c.name, c.address, err)
		return err
	}

	c.conn = conn
	c.logger.LogInfof("%s: connected to gRPC server at %s", c.name, c.address)
	return nil
}

func (c *StreamClient[T]) increaseBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > c.maxReconnectDelay {
		return c.maxReconnectDelay
	}
	return next
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func NewEventStreamClient(
	address string,
	filter *EventFilter,
	logger log.Logger,
	opts ...grpc.DialOption,
) *StreamClient[*Event] {
	return NewStreamClient[*Event](
		"event",
		address,
		logger,
		func(ctx context.Context, conn *grpc.ClientConn) (func() (*Event, error), error) {
			cli := NewEventServiceClient(conn)
			stream, err := cli.StreamEvents(ctx, &EventStreamRequest{Filter: filter})
			if err != nil {
				return nil, err
			}
			return stream.Recv, nil
		},
		opts...,
	)
}

func NewTransactionStreamClient(
	address string,
	filter *TransactionFilter,
	logger log.Logger,
	opts ...grpc.DialOption,
) *StreamClient[*Transaction] {
	return NewStreamClient[*Transaction](
		"transaction",
		address,
		logger,
		func(ctx context.Context, conn *grpc.ClientConn) (func() (*Transaction, error), error) {
			cli := NewTransactionServiceClient(conn)
			stream, err := cli.StreamTransactions(ctx, &TransactionStreamRequest{Filter: filter})
			if err != nil {
				return nil, err
			}
			return stream.Recv, nil
		},
		opts...,
	)
}
