package iotaconn

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/iotaledger/hive.go/log"
)

type WebsocketClient struct {
	idCounter         uint32
	url               string
	conn              *websocket.Conn
	writeQueue        chan *jsonrpcMessage
	readers           sync.Map // id -> chan *jsonrpcMessage
	log               log.Logger
	shutdownWaitGroup sync.WaitGroup
	reconnectMx       sync.Mutex
	subscriptions     []*subscription
	pendingCalls      sync.Map
}

func NewWebsocketClient(
	ctx context.Context,
	url string,
	log log.Logger,
) (*WebsocketClient, error) {
	c := &WebsocketClient{
		url:           url,
		writeQueue:    make(chan *jsonrpcMessage),
		log:           log,
		subscriptions: make([]*subscription, 0, 2),
	}

	err := c.reconnect(ctx)
	if err != nil {
		return nil, err
	}

	c.shutdownWaitGroup.Add(1)
	go c.loop(ctx)
	return c, nil
}

func (c *WebsocketClient) WaitUntilStopped() {
	c.shutdownWaitGroup.Wait()
}

func (c *WebsocketClient) loop(ctx context.Context) {
	defer c.shutdownWaitGroup.Done()

	type readMsgResult struct {
		messageType int
		p           []byte
	}
	receivedMsgs := make(chan readMsgResult)
	go func() {
		c.log.LogInfof("websocket loop started")
		defer c.log.LogInfof("websocket loop finished")
		defer close(receivedMsgs)
		for {
			m, p, err := c.readMessage()
			if err != nil {
				c.log.LogErrorf("WebsocketClient read loop: %s", err)
				continue
			}
			var j *jsonrpcMessage
			if err := json.Unmarshal(p, &j); err != nil {
				c.log.LogErrorf("WebsocketClient: could not unmarshal response body: %s", err)
				continue
			}
			c.log.LogDebugf("ws message was read: %v, %v", j.ID, j.Method)
			receivedMsgs <- readMsgResult{messageType: m, p: p}
		}
	}()

	defer c.conn.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case msgToSend := <-c.writeQueue:
			reqBody, err := json.Marshal(msgToSend)
			if err != nil {
				c.log.LogErrorf("WebsocketClient: could not marshal json: %s", err)
				continue
			}
			err = c.writeMessage(websocket.TextMessage, reqBody)
			if err != nil {
				c.log.LogErrorf("WebsocketClient: write error: %s", err)
				return
			}
		case receivedMsg, ok := <-receivedMsgs:
			if !ok {
				return
			}

			switch receivedMsg.messageType {
			case websocket.TextMessage:
				var m *jsonrpcMessage
				if err := json.Unmarshal(receivedMsg.p, &m); err != nil {
					c.log.LogErrorf("WebsocketClient: could not unmarshal response body: %s", err)
					continue
				}
				var id string
				if len(m.ID) > 0 {
					// this is a response to a method call
					id = string(m.ID)
					c.log.LogDebugf("response to method call: %+v", m.ID)
				} else if m.Method != "" {
					// this is a subscription message
					var s struct {
						Subscription uint64 `json:"subscription"`
					}
					if err := json.Unmarshal(m.Params, &s); err != nil {
						c.log.LogErrorf("WebsocketClient: could not unmarshal subscription params: %s", err)
						continue
					}
					id = fmt.Sprintf("%s:%d", m.Method, s.Subscription)
					c.log.LogDebugf("subscription message: %v", id)
				} else {
					c.log.LogErrorf("WebsocketClient: cannot identify message: %s", receivedMsg.p)
					continue
				}
				readCh, ok := c.readers.Load(id)
				if ok {
					readCh.(chan *jsonrpcMessage) <- m
				} else {
					// this can sometimes happen, but it's not an issue: the channel should be associated with the new id by now
					c.log.LogErrorf("WebsocketClient: no reader for message: %s", receivedMsg.p)
					continue
				}

			default:
				c.log.LogWarnf("WebsocketClient: ignoring binary message: %x", receivedMsg.p)
			}
		}
	}
}

func (c *WebsocketClient) readMessage() (messageType int, p []byte, err error) {
	if c.conn == nil {
		return 0, nil, fmt.Errorf("connection is nil")
	}

	messageType, p, err = c.conn.ReadMessage()
	if err != nil {
		c.log.LogWarnf("read failed: %s", err)
		if reconnErr := c.reconnect(context.Background()); reconnErr != nil {
			return 0, nil, fmt.Errorf("read failed and reconnect failed: %w", err)
		}
		return c.readMessage()
	}
	return messageType, p, nil
}

func (c *WebsocketClient) writeMessage(messageType int, data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	err := c.conn.WriteMessage(messageType, data)
	if err != nil {
		c.log.LogWarnf("write failed: %s", err)
		if reconnErr := c.reconnect(context.Background()); reconnErr != nil {
			return fmt.Errorf("write failed and reconnect failed: %w", err)
		}
		return c.writeMessage(messageType, data)
	}
	return nil
}

func (c *WebsocketClient) writeMsg(method JsonRPCMethod, args ...interface{}) (string, error) {
	msg, err := c.newMessage(method.String(), args...)
	if err != nil {
		return "", err
	}
	id := string(msg.ID)
	readCh := make(chan *jsonrpcMessage)
	c.readers.Store(id, readCh)
	c.writeQueue <- msg
	return id, nil
}

type subscription struct {
	method JsonRPCMethod
	args   []interface{}
	id     string
	uuid   uuid.UUID
}

type call struct {
	method JsonRPCMethod
	args   []interface{}
	id     string
}

func (c *WebsocketClient) CallContext(
	ctx context.Context,
	result interface{},
	method JsonRPCMethod,
	args ...interface{},
) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}

	id, err := c.writeMsg(method, args...)
	if err != nil {
		return err
	}

	c.pendingCalls.Store(id, &call{method: method, args: args, id: id})
	defer func() {
		c.pendingCalls.Delete(id)
	}()

	readCh, _ := c.readers.Load(id)
	defer c.readers.Delete(id)
	c.log.LogDebugf("waiting for response to %s", id)
	respmsg := <-readCh.(chan *jsonrpcMessage)
	c.log.LogDebugf("response to %s received", id)
	if respmsg.Error != nil {
		return respmsg.Error
	}
	if len(respmsg.Result) == 0 {
		return ErrNoResult
	}
	return json.Unmarshal(respmsg.Result, result)
}

func (c *WebsocketClient) Subscribe(
	ctx context.Context,
	resultCh chan<- []byte,
	method JsonRPCMethod,
	args ...interface{},
) error {
	var subID uint64
	err := c.CallContext(ctx, &subID, method, args...)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s:%d", method, subID)
	readCh := make(chan *jsonrpcMessage)
	c.readers.Store(id, readCh)

	c.subscriptions = append(
		c.subscriptions, &subscription{
			method: method,
			args:   args,
			id:     id,
			uuid:   uuid.New(),
		},
	)
	c.log.LogDebugf("subscribing to %s", method)

	go func() {
		defer close(resultCh)
		defer c.readers.Delete(id)
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-readCh:
				if msg.Error != nil {
					c.log.LogErrorf("subscription error: %s", msg.Error)
					return
				}
				if len(msg.Params) == 0 {
					c.log.LogWarnf("Ignoring websocket subscription message: %+v\n", msg)
					continue
				}
				var params jsonrpcWebsocketParams
				if err := json.Unmarshal(msg.Params, &params); err != nil {
					c.log.LogErrorf("could not unmarshal msg.Params: %s", err)
					continue
				}
				c.log.LogDebugf("subscription result: %+v", params.Result)
				resultCh <- params.Result
			}
		}
	}()

	return nil
}

func (c *WebsocketClient) newMessage(method string, paramsIn ...interface{}) (*jsonrpcMessage, error) {
	id := c.nextID()
	msg := &jsonrpcMessage{
		Version: version,
		ID:      json.RawMessage(id),
		Method:  method,
	}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		if msg.Params, err = json.Marshal(paramsIn); err != nil {
			return nil, err
		}
	}
	return msg, nil
}

func (c *WebsocketClient) nextID() string {
	id := atomic.AddUint32(&c.idCounter, 1)
	return strconv.FormatUint(uint64(id), 10)
}

func (c *WebsocketClient) reconnect(ctx context.Context) error {
	c.log.LogDebugf("reconnecting")
	if c.reconnectMx.TryLock() {
		defer c.reconnectMx.Unlock()
	} else {
		// already reconnecting, try again later
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	if c.conn != nil {
		c.conn.Close()
	}

	const retryInterval = time.Second
	attempt := 1

	for {
		dialer := websocket.Dialer{}
		conn, _, err := dialer.DialContext(ctx, c.url, nil)
		if err != nil {
			c.log.LogWarnf("connection attempt %d failed: %v", attempt, err)
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while reconnecting: %w", ctx.Err())
			case <-time.After(retryInterval):
				attempt++
				continue
			}
		}

		c.conn = conn
		c.log.LogDebugf("new connection set after %d attempts", attempt)

		// recreating subscriptions and recreating pending calls. This should happen asynchronously because it needs the loop to be running
		go c.resubscribe(ctx)
		go c.recreatePendingCalls()

		return nil
	}
}

// recreatePendingCalls recreates pending calls. Errors in this function will cause particular calls to not complete, so no need to fail other calls
func (c *WebsocketClient) recreatePendingCalls() {
	c.pendingCalls.Range(
		func(key, value interface{}) bool {
			call := value.(*call)
			oldId := key.(string)

			msg, err := c.newMessage(call.method.String(), call.args...)
			if err != nil {
				c.log.LogErrorf("failed to recreate pending call %s: %s", oldId, err)
				return true
			}

			newId := string(msg.ID)

			c.log.LogDebugf("recreate writing message: oldId: %s, newId: %s, %+v", oldId, newId, msg)

			ch, ok := c.readers.Load(oldId)
			if !ok {
				c.log.LogErrorf("failed to recreate pending call: reader for old id %s not found", oldId)
				return true
			}
			readCh := ch.(chan *jsonrpcMessage)
			c.readers.Store(newId, readCh)
			c.writeQueue <- msg

			c.readers.Delete(oldId)

			return true
		},
	)
}

// resubscribe to subscriptions. Errors in this function probably mean that subscription configurations themself contain errors, so ignoring
func (c *WebsocketClient) resubscribe(ctx context.Context) {
	c.log.LogDebugf("resubscribing to %d subscriptions", len(c.subscriptions))
	defer c.log.LogDebugf("resubscribed")

	for _, sub := range c.subscriptions {
		c.log.LogDebugf("resubscribing to %s, %+v", sub.method, sub.args)
		defer c.log.LogDebugf("resubscribed to %s", sub.method)

		method := sub.method
		args := sub.args
		oldId := sub.id

		var subID uint64
		err := c.CallContext(ctx, &subID, method, args...)
		if err != nil {
			c.log.LogErrorf("failed to resubscribe to %s: %s", method, err)
			continue
		}
		newId := fmt.Sprintf("%s:%d", method, subID)

		// store reader channel with new id
		ch, ok := c.readers.Load(oldId)
		c.readers.Delete(oldId)
		if !ok {
			c.log.LogErrorf("reader for old id %s not found", oldId)
			continue
		}
		c.readers.Store(newId, ch)

		// need to update subscription id so that next resubscribe works
		sub.id = newId
	}
}
