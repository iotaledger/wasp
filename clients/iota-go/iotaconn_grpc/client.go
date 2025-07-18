package iotaconn_grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// EventStreamClient wraps the gRPC client with automatic reconnection
type EventStreamClient struct {
	address     string
	dialOptions []grpc.DialOption
	conn        *grpc.ClientConn
	client      EventServiceClient

	mu          sync.RWMutex
	subscribers map[string]*subscription

	// Reconnection settings
	reconnectInterval time.Duration
	maxReconnectDelay time.Duration

	// Context for shutting down
	ctx    context.Context
	cancel context.CancelFunc
}

type subscription struct {
	id     string
	filter *EventFilter
	events chan *Event
	errors chan error
	ctx    context.Context
	cancel context.CancelFunc
}

// NewEventStreamClient creates a new client with automatic reconnection
func NewEventStreamClient(address string, opts ...grpc.DialOption) *EventStreamClient {
	ctx, cancel := context.WithCancel(context.Background())

	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                10 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			},
		),
	}

	// Merge provided options with defaults
	allOpts := append(defaultOpts, opts...)

	return &EventStreamClient{
		address:           address,
		dialOptions:       allOpts,
		subscribers:       make(map[string]*subscription),
		reconnectInterval: time.Second,
		maxReconnectDelay: 30 * time.Second,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Connect establishes connection to the gRPC server
func (c *EventStreamClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	conn, err := grpc.DialContext(c.ctx, c.address, c.dialOptions...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = NewEventServiceClient(conn)

	log.Printf("Connected to gRPC server at %s", c.address)
	return nil
}

// SubscribeEvents subscribes to events matching the filter and returns channels for events and errors
func (c *EventStreamClient) SubscribeEvents(filter *EventFilter) (<-chan *Event, <-chan error, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate unique subscription ID
	subID := generateSubscriptionID()

	// Create subscription context
	subCtx, subCancel := context.WithCancel(c.ctx)

	sub := &subscription{
		id:     subID,
		filter: filter,
		events: make(chan *Event, 100), // Buffered channel
		errors: make(chan error, 10),
		ctx:    subCtx,
		cancel: subCancel,
	}

	c.subscribers[subID] = sub

	// Start the streaming goroutine
	go c.streamEvents(sub)

	// Return channels and unsubscribe function
	unsubscribe := func() {
		c.unsubscribe(subID)
	}

	return sub.events, sub.errors, unsubscribe
}

// streamEvents handles the streaming for a specific subscription
func (c *EventStreamClient) streamEvents(sub *subscription) {
	defer func() {
		close(sub.events)
		close(sub.errors)
	}()

	backoff := c.reconnectInterval

	for {
		select {
		case <-sub.ctx.Done():
			return
		default:
		}

		// Ensure we have a connection
		if err := c.ensureConnection(); err != nil {
			sub.errors <- err
			c.sleep(backoff)
			backoff = c.increaseBackoff(backoff)
			continue
		}

		// Create stream request
		req := &EventStreamRequest{
			Filter: sub.filter,
		}

		// Start streaming
		stream, err := c.client.StreamEvents(sub.ctx, req)
		if err != nil {
			sub.errors <- err
			c.sleep(backoff)
			backoff = c.increaseBackoff(backoff)
			continue
		}

		// Reset backoff on successful connection
		backoff = c.reconnectInterval
		h, err := stream.Header()

		fmt.Println(h)
		fmt.Println(err)

		// Process events from stream
		for {
			select {
			case <-sub.ctx.Done():
				return
			default:
			}

			event, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Stream ended for subscription %s", sub.id)
				break // Will reconnect
			}
			if err != nil {
				log.Printf("Stream error for subscription %s: %v", sub.id, err)
				sub.errors <- err
				break // Will reconnect
			}

			// Send event to channel (non-blocking)
			select {
			case sub.events <- event:
			case <-sub.ctx.Done():
				return
			default:
				// Channel is full, drop oldest event
				log.Printf("Event channel full for subscription %s, dropping event", sub.id)
				select {
				case <-sub.events:
				default:
				}
				select {
				case sub.events <- event:
				default:
				}
			}
		}

		// Wait before reconnecting
		c.sleep(time.Second)
	}
}

// ensureConnection checks if connection is healthy and reconnects if needed
func (c *EventStreamClient) ensureConnection() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return c.connectLocked()
	}

	state := c.conn.GetState()
	if state == connectivity.TransientFailure || state == connectivity.Shutdown {
		log.Printf("Connection state: %v, reconnecting...", state)
		return c.connectLocked()
	}

	return nil
}

// connectLocked connects to server (must be called with mutex held)
func (c *EventStreamClient) connectLocked() error {
	if c.conn != nil {
		c.conn.Close()
	}

	conn, err := grpc.DialContext(c.ctx, c.address, c.dialOptions...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = NewEventServiceClient(conn)

	log.Printf("Reconnected to gRPC server at %s", c.address)
	return nil
}

// unsubscribe removes a subscription
func (c *EventStreamClient) unsubscribe(subID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if sub, exists := c.subscribers[subID]; exists {
		sub.cancel()
		delete(c.subscribers, subID)
		log.Printf("Unsubscribed subscription %s", subID)
	}
}

// Close closes all subscriptions and the connection
func (c *EventStreamClient) Close() error {
	c.cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel all subscriptions
	for id, sub := range c.subscribers {
		sub.cancel()
		delete(c.subscribers, id)
	}

	// Close connection
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}

	return nil
}

// Helper methods
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

func generateSubscriptionID() string {
	return fmt.Sprintf("sub_%d", time.Now().UnixNano())
}
