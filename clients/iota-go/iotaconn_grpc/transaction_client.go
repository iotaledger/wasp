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

type TransactionStreamClient struct {
	address string
	filter  *TransactionFilter
	options []grpc.DialOption
	logger  log.Logger

	conn   *grpc.ClientConn
	client TransactionServiceClient

	events chan *Transaction
	wg     sync.WaitGroup

	reconnectInterval time.Duration
	maxReconnectDelay time.Duration
}

func NewTransactionStreamClient(
	address string,
	filter *TransactionFilter,
	logger log.Logger,
	opts ...grpc.DialOption,
) *TransactionStreamClient {
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
	return &TransactionStreamClient{
		address:           address,
		filter:            filter,
		options:           all,
		logger:            logger,
		events:            make(chan *Transaction, defaultBuf),
		reconnectInterval: 5 * time.Second,
		maxReconnectDelay: 30 * time.Second,
	}
}

func (c *TransactionStreamClient) Start(ctx context.Context) <-chan *Transaction {
	c.wg.Add(1)
	go c.run(ctx)
	return c.events
}

func (c *TransactionStreamClient) WaitUntilStopped() {
	c.wg.Wait()
}

func (c *TransactionStreamClient) run(ctx context.Context) {
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

		stream, err := c.client.StreamTransactions(ctx, &TransactionStreamRequest{Filter: c.filter})
		if err != nil {
			c.logger.LogErrorf("Failed to create stream: %v", err)
			if !sleepCtx(ctx, backoff) {
				break
			}
			backoff = c.increaseBackoff(backoff)
			continue
		}

		c.logger.LogInfof("Transaction stream established")
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

func (c *TransactionStreamClient) dial() error {
	if c.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(c.address, c.options...)
	if err != nil {
		c.logger.LogErrorf("Failed to connect to %s: %v", c.address, err)
		return err
	}

	c.conn = conn
	c.client = NewTransactionServiceClient(conn)
	c.logger.LogInfof("Connected to gRPC server at %s", c.address)
	return nil
}

func (c *TransactionStreamClient) increaseBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > c.maxReconnectDelay {
		return c.maxReconnectDelay
	}
	return next
}
