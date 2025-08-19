package iotaconn_grpc

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type EventStreamClient struct {
	address string
	filter  *EventFilter
	options []grpc.DialOption

	conn   *grpc.ClientConn
	client EventServiceClient

	events chan *Event

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
	mu sync.RWMutex

	reconnectInterval time.Duration
	maxReconnectDelay time.Duration
}

func NewEventStreamClient(
	ctx context.Context, address string, filter *EventFilter,
	opts ...grpc.DialOption,
) *EventStreamClient {
	ctx, cancel := context.WithCancel(ctx)

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

	allOpts := append(defaultOpts, opts...)

	return &EventStreamClient{
		address:           address,
		filter:            filter,
		options:           allOpts,
		events:            make(chan *Event, 100),
		ctx:               ctx,
		cancel:            cancel,
		reconnectInterval: time.Second,
		maxReconnectDelay: 30 * time.Second,
	}
}

func (c *EventStreamClient) Start() <-chan *Event {
	c.wg.Add(1)
	go c.streamLoop()
	return c.events
}

func (c *EventStreamClient) streamLoop() {
	defer c.wg.Done()
	defer close(c.events)

	backoff := c.reconnectInterval
	log.Printf("Starting event stream client for %s", c.address)

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Event stream client shutting down")
			return
		default:
		}

		if err := c.ensureConnected(); err != nil {
			c.sleep(backoff)
			backoff = c.increaseBackoff(backoff)
			continue
		}

		backoff = c.reconnectInterval

		if c.processStream() {
			c.sleep(c.reconnectInterval)
		}
	}
}

func (c *EventStreamClient) processStream() bool {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return false
	}

	req := &EventStreamRequest{
		Filter: c.filter,
	}

	stream, err := client.StreamEvents(c.ctx, req)
	if err != nil {
		log.Printf("Failed to create stream: %v", err)
		return false
	}

	log.Printf("Event stream established")

	for {
		select {
		case <-c.ctx.Done():
			return true
		default:
		}

		event, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Stream ended normally")
			return true
		}
		if err != nil {
			log.Printf("Stream receive error: %v", err)
			return false
		}

		select {
		case c.events <- event:
		case <-c.ctx.Done():
			return true
		default:
			// Channel is full, drop the event
			log.Printf("Event channel full, dropping event")
		}
	}
}

func (c *EventStreamClient) ensureConnected() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to connect/reconnect
	if c.conn == nil || c.isConnectionUnhealthy() {
		return c.connectLocked()
	}

	return nil
}

func (c *EventStreamClient) isConnectionUnhealthy() bool {
	if c.conn == nil {
		return true
	}

	state := c.conn.GetState()
	return state == connectivity.TransientFailure ||
		state == connectivity.Shutdown
}

func (c *EventStreamClient) connectLocked() error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}

	conn, err := grpc.NewClient(c.address, c.options...)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", c.address, err)
		return err
	}

	c.conn = conn
	c.client = NewEventServiceClient(conn)

	log.Printf("Connected to gRPC server at %s", c.address)
	return nil
}

func (c *EventStreamClient) sleep(duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-c.ctx.Done():
	}
}

func (c *EventStreamClient) increaseBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > c.maxReconnectDelay {
		return c.maxReconnectDelay
	}
	return next
}

func (c *EventStreamClient) Close() error {
	log.Printf("Closing event stream client")

	c.cancel()
	c.wg.Wait()
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.client = nil
		return err
	}

	log.Printf("Event stream client closed")
	return nil
}

func (c *EventStreamClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

func (c *EventStreamClient) GetConnectionState() connectivity.State {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return connectivity.Shutdown
	}

	return c.conn.GetState()
}
