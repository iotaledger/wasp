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

// EventStreamClient provides a simple, reconnecting gRPC event stream client
type EventStreamClient struct {
	address string
	filter  *EventFilter
	options []grpc.DialOption

	conn   *grpc.ClientConn
	client EventServiceClient

	events chan *Event
	errors chan error

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
	mu sync.RWMutex

	// Reconnection settings
	reconnectInterval time.Duration
	maxReconnectDelay time.Duration
}

// NewEventStreamClient creates a new event stream client
func NewEventStreamClient(address string, filter *EventFilter, opts ...grpc.DialOption) *EventStreamClient {
	ctx, cancel := context.WithCancel(context.Background())

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

	// Merge provided options with defaults
	allOpts := append(defaultOpts, opts...)

	return &EventStreamClient{
		address:           address,
		filter:            filter,
		options:           allOpts,
		events:            make(chan *Event, 100),
		errors:            make(chan error, 10),
		ctx:               ctx,
		cancel:            cancel,
		reconnectInterval: time.Second,
		maxReconnectDelay: 30 * time.Second,
	}
}

// Start begins the event streaming and returns channels for events and errors
func (c *EventStreamClient) Start() (<-chan *Event, <-chan error) {
	c.wg.Add(1)
	go c.streamLoop()
	return c.events, c.errors
}

// streamLoop is the main event streaming loop with reconnection logic
func (c *EventStreamClient) streamLoop() {
	defer c.wg.Done()
	defer close(c.events)
	defer close(c.errors)

	backoff := c.reconnectInterval
	log.Printf("Starting event stream client for %s", c.address)

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("Event stream client shutting down")
			return
		default:
		}

		// Ensure we have a healthy connection
		if err := c.ensureConnected(); err != nil {
			c.sendError(err)
			c.sleep(backoff)
			backoff = c.increaseBackoff(backoff)
			continue
		}

		// Reset backoff on successful connection
		backoff = c.reconnectInterval

		// Process the stream
		if c.processStream() {
			// Stream ended normally, wait before reconnecting
			c.sleep(c.reconnectInterval)
		}
	}
}

// processStream handles a single stream connection
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
		c.sendError(err)
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
			c.sendError(err)
			return false
		}

		// Send event to channel (non-blocking with drop policy)
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

// ensureConnected ensures we have a healthy gRPC connection
func (c *EventStreamClient) ensureConnected() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to connect/reconnect
	if c.conn == nil || c.isConnectionUnhealthy() {
		return c.connectLocked()
	}

	return nil
}

// isConnectionUnhealthy checks if the current connection is unhealthy
func (c *EventStreamClient) isConnectionUnhealthy() bool {
	if c.conn == nil {
		return true
	}

	state := c.conn.GetState()
	return state == connectivity.TransientFailure ||
		state == connectivity.Shutdown
}

// connectLocked establishes a new connection (must be called with mutex held)
func (c *EventStreamClient) connectLocked() error {
	// Close existing connection
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}

	// Create new connection
	conn, err := grpc.DialContext(c.ctx, c.address, c.options...)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", c.address, err)
		return err
	}

	c.conn = conn
	c.client = NewEventServiceClient(conn)

	log.Printf("Connected to gRPC server at %s", c.address)
	return nil
}

// sendError sends an error to the error channel (non-blocking)
func (c *EventStreamClient) sendError(err error) {
	select {
	case c.errors <- err:
	default:
		// Error channel is full, drop the error
		log.Printf("Error channel full, dropping error: %v", err)
	}
}

// sleep waits for the specified duration or until context is cancelled
func (c *EventStreamClient) sleep(duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-c.ctx.Done():
	}
}

// increaseBackoff implements exponential backoff with maximum delay
func (c *EventStreamClient) increaseBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > c.maxReconnectDelay {
		return c.maxReconnectDelay
	}
	return next
}

// Close gracefully shuts down the client
func (c *EventStreamClient) Close() error {
	log.Printf("Closing event stream client")

	// Cancel context to stop all operations
	c.cancel()

	// Wait for stream loop to finish
	c.wg.Wait()

	// Close connection
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

// IsConnected returns whether the client is currently connected
func (c *EventStreamClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// GetConnectionState returns the current connection state
func (c *EventStreamClient) GetConnectionState() connectivity.State {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return connectivity.Shutdown
	}

	return c.conn.GetState()
}
