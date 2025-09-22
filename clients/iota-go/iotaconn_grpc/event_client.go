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

type EventStreamClient struct {
	address string
	filter  *EventFilter
	options []grpc.DialOption
	logger  log.Logger

	conn   *grpc.ClientConn
	client EventServiceClient

	events chan *Event
	wg     sync.WaitGroup

	reconnectInterval time.Duration
	maxReconnectDelay time.Duration
}

const defaultBuf = 64

func NewEventStreamClient(
	address string,
	filter *EventFilter,
	logger log.Logger,
	opts ...grpc.DialOption,
) *EventStreamClient {
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
	return &EventStreamClient{
		address:           address,
		filter:            filter,
		options:           all,
		logger:            logger,
		events:            make(chan *Event, defaultBuf),
		reconnectInterval: 5 * time.Second,
		maxReconnectDelay: 30 * time.Second,
	}
}

func (c *EventStreamClient) Start(ctx context.Context) <-chan *Event {
	c.wg.Add(1)
	go c.run(ctx)
	return c.events
}

func (c *EventStreamClient) WaitUntilStopped() {
	c.wg.Wait()
}

func (c *EventStreamClient) run(ctx context.Context) {
	defer c.wg.Done()
	defer close(c.events)

	if err := c.dial(); err != nil {
		c.logger.LogErrorf("gRPC dial failed: %v", err)
	}

	backoff := c.reconnectInterval

	for {
		if ctx.Err() != nil {
			break
		}

		if c.client == nil {
			if !sleepCtx(ctx, backoff) {
				break
			}
			backoff = c.increaseBackoff(backoff)
			_ = c.dial()
			continue
		}

		stream, err := c.client.StreamEvents(ctx, &EventStreamRequest{Filter: c.filter})
		if err != nil {
			c.logger.LogErrorf("Failed to create stream: %v", err)
			if !sleepCtx(ctx, backoff) {
				break
			}
			backoff = c.increaseBackoff(backoff)
			continue
		}

		c.logger.LogInfof("Event stream established")
		backoff = c.reconnectInterval

		for {
			if ctx.Err() != nil {
				break
			}
			evt, err := stream.Recv()
			if err == io.EOF {
				c.logger.LogErrorf("Stream ended")
				break
			}
			if err != nil {
				c.logger.LogErrorf("Stream receive error: %v", err)
				break
			}

			select {
			case c.events <- evt:
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
	c.client = nil
	if conn != nil {
		_ = conn.Close()
	}
}

func (c *EventStreamClient) dial() error {
	if c.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(c.address, c.options...)
	if err != nil {
		c.logger.LogErrorf("Failed to connect to %s: %v", c.address, err)
		return err
	}

	c.conn = conn
	c.client = NewEventServiceClient(conn)
	c.logger.LogInfof("Connected to gRPC server at %s", c.address)
	return nil
}

func (c *EventStreamClient) increaseBackoff(current time.Duration) time.Duration {
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
