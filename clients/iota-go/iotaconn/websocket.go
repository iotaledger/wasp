package iotaconn

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"

	"github.com/iotaledger/hive.go/logger"
)

type WebsocketClient struct {
	idCounter         uint32
	conn              *websocket.Conn
	writeQueue        chan *jsonrpcMessage
	readers           sync.Map // id -> chan *jsonrpcMessage
	log               *logger.Logger
	shutdownWaitGroup sync.WaitGroup
	url               string
	ctx               context.Context
	disconnected      chan struct{}
}

func NewWebsocketClient(
	ctx context.Context,
	url string,
	log *logger.Logger,
) (*WebsocketClient, error) {
	c := &WebsocketClient{
		url:          url,
		ctx:          ctx,
		writeQueue:   make(chan *jsonrpcMessage),
		log:          log,
		disconnected: make(chan struct{}),
	}

	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(c.ctx, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket server: %w", err)
	}

	c.conn = conn

	c.shutdownWaitGroup.Add(1)
	go c.loop(ctx)
	return c, nil
}

func (c *WebsocketClient) WaitUntilStopped() {
	c.shutdownWaitGroup.Wait()
}

func (c *WebsocketClient) loop(ctx context.Context) {
	defer c.shutdownWaitGroup.Done()

	if err := c.runLoop(ctx); err != nil {
		select {
		case <-ctx.Done():
			return
		default:
			close(c.disconnected)
			return
		}
	}
}

var counter int

func (c *WebsocketClient) runLoop(ctx context.Context) error {
	type readMsgResult struct {
		messageType int
		p           []byte
	}
	receivedMsgs := make(chan readMsgResult)
	go func() {
		defer close(receivedMsgs)
		for {
			m, p, err := c.conn.ReadMessage()
			if err != nil {
				c.log.Errorf("WebsocketClient read loop: %s", err)
				return
			} else {
				receivedMsgs <- readMsgResult{messageType: m, p: p}
			}
		}
	}()

	defer c.conn.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msgToSend := <-c.writeQueue:
			reqBody, err := json.Marshal(msgToSend)
			if err != nil {
				c.log.Errorf("WebsocketClient: could not marshal json: %s", err)
				continue
			}
			err = c.conn.WriteMessage(websocket.TextMessage, reqBody)
			if err != nil {
				c.log.Errorf("WebsocketClient: write error: %s", err)
				return fmt.Errorf("connection closed: %w", err)
			}
		case receivedMsg, ok := <-receivedMsgs:
			if !ok {
				return fmt.Errorf("connection closed")
			}
			counter++
			if counter%30 == 0 {
				c.log.Info("simulate connection loss")
				c.conn.Close()
			}
			switch receivedMsg.messageType {
			case websocket.TextMessage:
				var m *jsonrpcMessage
				if err := json.Unmarshal(receivedMsg.p, &m); err != nil {
					c.log.Errorf("WebsocketClient: could not unmarshal response body: %s", err)
					continue
				}
				var id string
				if len(m.ID) > 0 {
					// this is a response to a method call
					id = string(m.ID)
				} else if m.Method != "" {
					// this is a subscription message
					var s struct {
						Subscription uint64 `json:"subscription"`
					}
					if err := json.Unmarshal(m.Params, &s); err != nil {
						c.log.Errorf("WebsocketClient: could not unmarshal subscription params: %s", err)
						continue
					}
					id = fmt.Sprintf("%s:%d", m.Method, s.Subscription)
				} else {
					c.log.Errorf("WebsocketClient: cannot identify message: %s", receivedMsg.p)
					continue
				}
				readCh, ok := c.readers.Load(id)
				if ok {
					readCh.(chan *jsonrpcMessage) <- m
				} else {
					c.log.Errorf("WebsocketClient: no reader for message: %s", receivedMsg.p)
					continue
				}

			default:
				c.log.Warnf("WebsocketClient: ignoring binary message: %x", receivedMsg.p)
			}
		}
	}
}

func (c *WebsocketClient) writeMsg(ctx context.Context, method JsonRPCMethod, args ...interface{}) (string, error) {
	msg, err := c.newMessage(method.String(), args...)
	if err != nil {
		return "", err
	}
	id := string(msg.ID)
	readCh := make(chan *jsonrpcMessage)
	c.readers.Store(id, readCh)

	select {
	case <-ctx.Done():
		c.log.Info("writeMsg: context done")
		return "", ctx.Err()
	case c.writeQueue <- msg:
		return id, nil
	}
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
	id, err := c.writeMsg(ctx, method, args...)
	if err != nil {
		return err
	}
	readCh, _ := c.readers.Load(id)
	defer c.readers.Delete(id)

	var respmsg *jsonrpcMessage
	select {
	case <-ctx.Done():
		c.log.Info("CallContext: context done")
		return ctx.Err()
	case respmsg = <-readCh.(chan *jsonrpcMessage):
		if respmsg.Error != nil {
			return respmsg.Error
		}
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

	go func() {
		defer close(resultCh)
		defer c.readers.Delete(id)
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-readCh:
				if msg.Error != nil {
					c.log.Errorf("subscription error: %s", msg.Error)
					return
				}
				if len(msg.Params) == 0 {
					c.log.Warnf("Ignoring websocket subscription message: %+v\n", msg)
					continue
				}
				var params jsonrpcWebsocketParams
				if err := json.Unmarshal(msg.Params, &params); err != nil {
					c.log.Errorf("could not unmarshal msg.Params: %s", err)
					continue
				}
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

func (c *WebsocketClient) Disconnected() <-chan struct{} {
	return c.disconnected
}
