package suiconn

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	idCounter  uint32
	conn       *websocket.Conn
	writeQueue chan *jsonrpcMessage
	readers    sync.Map // id -> chan *jsonrpcMessage
}

type CallOp struct {
	Method string
	Params []interface{}
}

func NewWebsocketClient(ctx context.Context, url string) *WebsocketClient {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to websocket server: %s, %s", err, url))
	}
	c := &WebsocketClient{
		conn:       conn,
		writeQueue: make(chan *jsonrpcMessage),
	}
	go c.loop(ctx)
	return c
}

func (c *WebsocketClient) loop(ctx context.Context) {
	type readMsgResult struct {
		messageType int
		p           []byte
	}
	receivedMsgs := make(chan readMsgResult)
	go func() {
		for {
			m, p, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("WebsocketClient read loop: %s", err)
				return
			} else {
				receivedMsgs <- readMsgResult{messageType: m, p: p}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msgToSend := <-c.writeQueue:
			reqBody, err := json.Marshal(msgToSend)
			if err != nil {
				log.Printf("WebsocketClient: could not marshal json: %s", err)
				continue
			}
			err = c.conn.WriteMessage(websocket.TextMessage, reqBody)
			if nil != err {
				log.Printf("WebsocketClient: write error: %s", err)
				return
			}
		case receivedMsg := <-receivedMsgs:
			switch receivedMsg.messageType {
			case websocket.TextMessage:
				var m *jsonrpcMessage
				if err := json.Unmarshal(receivedMsg.p, &m); err != nil {
					log.Printf("WebsocketClient: could not unmarshal response body: %s", err)
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
						log.Printf("WebsocketClient: could not unmarshal subscription params: %s", err)
						continue
					}
					id = fmt.Sprintf("%s:%d", m.Method, s.Subscription)
				} else {
					log.Printf("WebsocketClient: cannot identify message: %s", receivedMsg.p)
					continue
				}
				readCh, ok := c.readers.Load(id)
				if ok {
					readCh.(chan *jsonrpcMessage) <- m
				} else {
					log.Printf("WebsocketClient: no reader for message: %s", receivedMsg.p)
					continue
				}

			default:
				log.Printf("WebsocketClient: ignoring binary message: %x", receivedMsg.p)
			}
		}
	}
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

func (c *WebsocketClient) CallContext(ctx context.Context, result interface{}, method JsonRPCMethod, args ...interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}
	id, err := c.writeMsg(method, args...)
	if err != nil {
		return err
	}
	readCh, _ := c.readers.Load(id)
	defer c.readers.Delete(id)
	respmsg := <-readCh.(chan *jsonrpcMessage)
	if respmsg.Error != nil {
		return respmsg.Error
	}
	if len(respmsg.Result) == 0 {
		return ErrNoResult
	}
	return json.Unmarshal(respmsg.Result, result)
}

func (c *WebsocketClient) Subscribe(ctx context.Context, resultCh chan<- []byte, method JsonRPCMethod, args ...interface{}) error {
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
					log.Printf("subscription error: %s", msg.Error)
					return
				}
				if len(msg.Params) == 0 {
					log.Printf("Ignoring websocket subscription message: %+v\n", msg)
					continue
				}
				var params jsonrpcWebsocketParams
				if err := json.Unmarshal(msg.Params, &params); err != nil {
					log.Fatalf("could not unmarshal msg.Params: %s", err)
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
