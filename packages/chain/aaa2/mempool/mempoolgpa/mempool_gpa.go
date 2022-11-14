package mempoolgpa

import (
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

const (
	shareReqNNodes   = 2
	shareReqInterval = 1 * time.Second
)

const (
	askMissingReqNNodes   = 2
	askMissingReqInterval = 300 * time.Millisecond
)

type scheduledMessages struct {
	messages []gpa.Message
	lastSent time.Time
}
type Impl struct {
	receiveRequests func(reqs ...isc.Request) []bool
	getRequest      func(id isc.RequestID) isc.Request
	committeeNodes  []gpa.NodeID
	accessNodes     []gpa.NodeID
	peersMutex      sync.RWMutex
	log             *logger.Logger

	missingReqMsgs      map[isc.RequestID]*scheduledMessages
	missingReqMsgsMutex sync.Mutex
	shareReqMsgs        map[isc.RequestID]*scheduledMessages
	shareReqMsgsMutex   sync.Mutex
}

var _ gpa.GPA = &Impl{}

func New(
	receiveRequests func(reqs ...isc.Request) []bool,
	getRequest func(id isc.RequestID) isc.Request,
	log *logger.Logger,
) *Impl {
	return &Impl{
		receiveRequests: receiveRequests,
		getRequest:      getRequest,
		log:             log,
		missingReqMsgs:  make(map[isc.RequestID]*scheduledMessages),
		shareReqMsgs:    make(map[isc.RequestID]*scheduledMessages),
	}
}

func (m *Impl) SetPeers(committeeNodes, accessNodes []gpa.NodeID) {
	m.peersMutex.Lock()
	defer m.peersMutex.Unlock()
	m.committeeNodes = committeeNodes
	m.accessNodes = accessNodes
}

type RemovedFromMempool struct {
	RequestIDs []isc.RequestID
}

func (m *Impl) Input(input gpa.Input) gpa.OutMessages {
	switch inp := input.(type) {
	case time.Time:
		return m.handleInputTick(inp)
	case RemovedFromMempool:
		return m.handleRemovedFromMempool(inp)
	case isc.Request:
		return m.handleInputRequest(inp)
	case *isc.RequestRef:
		return m.handleInputRequestRef(inp)
	default:
		m.log.Warnf("unexpected input %T: %+v", input, input)
	}
	return nil
}

func (m *Impl) handleInputTick(t time.Time) gpa.OutMessages {
	msgs := gpa.NoMessages()
	msgs.AddMany(m.nextMissingRequestMsgs(t))
	msgs.AddMany(m.nextShareRequestMsgs(t))
	return msgs
}

func (m *Impl) nextMissingRequestMsgs(t time.Time) []gpa.Message {
	return nextScheduledMsgs(
		&m.missingReqMsgsMutex,
		m.missingReqMsgs,
		askMissingReqNNodes,
		askMissingReqInterval,
		t,
	)
}

func (m *Impl) nextShareRequestMsgs(t time.Time) []gpa.Message {
	return nextScheduledMsgs(
		&m.shareReqMsgsMutex,
		m.shareReqMsgs,
		shareReqNNodes,
		shareReqInterval,
		t,
	)
}

func nextScheduledMsgs(
	mutex *sync.Mutex,
	msgsMap map[isc.RequestID]*scheduledMessages,
	nMessages int,
	interval time.Duration,
	t time.Time,
) []gpa.Message {
	mutex.Lock()
	defer mutex.Unlock()

	ret := []gpa.Message{}
	for reqid, msgs := range msgsMap {
		if t.Before(msgs.lastSent.Add(interval)) {
			// only add messages to be sent if the interval has passed
			continue
		}
		if len(msgs.messages) >= nMessages {
			ret = append(ret, msgs.messages...)
			delete(msgsMap, reqid)
		} else {
			msgs.lastSent = t
			ret = append(ret, msgs.messages[nMessages:]...)
			msgs.messages = msgs.messages[nMessages:]
		}
	}
	return ret
}

func (m *Impl) handleRemovedFromMempool(r RemovedFromMempool) gpa.OutMessages {
	m.shareReqMsgsMutex.Lock()
	defer m.shareReqMsgsMutex.Unlock()
	for _, rid := range r.RequestIDs {
		delete(m.shareReqMsgs, rid)
	}
	return nil
}

func (m *Impl) handleInputRequest(req isc.Request) gpa.OutMessages {
	// empty node ID because this request is coming from webapi
	return m.newShareRequestMessages(req, gpa.NodeID(""))
}

func (m *Impl) handleInputRequestRef(ref *isc.RequestRef) gpa.OutMessages {
	m.peersMutex.RLock()
	defer m.peersMutex.RUnlock()
	msgs := make([]gpa.Message, len(m.committeeNodes))
	for i, nodeID := range m.committeeNodes {
		msgs[i] = newMsgMissingRequest(ref, nodeID)
	}
	if len(msgs) <= askMissingReqNNodes {
		return gpa.NoMessages().AddMany(msgs)
	}
	// send the first iteration of messages right away
	ret := gpa.NoMessages().AddMany(msgs[:askMissingReqNNodes])
	// save the rest of the messages to be sent later
	msgs = msgs[askMissingReqNNodes:]
	if len(msgs) > 0 {
		m.missingReqMsgsMutex.Lock()
		m.missingReqMsgs[ref.ID] = &scheduledMessages{
			messages: msgs,
			lastSent: time.Now(),
		}
		m.missingReqMsgsMutex.Unlock()
	}
	return ret
}

func (m *Impl) newShareRequestMessages(req isc.Request, receivedFrom gpa.NodeID) gpa.OutMessages {
	m.peersMutex.RLock()
	defer m.peersMutex.RUnlock()
	// share to committee and access nodes
	allNodes := m.committeeNodes
	allNodes = append(allNodes, m.accessNodes...)
	msgs := []gpa.Message{}
	for _, nodeID := range allNodes {
		if nodeID != receivedFrom {
			msgs = append(msgs, newMsgShareRequest(req, true, nodeID))
		}
	}
	if len(msgs) <= shareReqNNodes {
		return gpa.NoMessages().AddMany(msgs)
	}
	// send the first iteration of messages right away
	ret := gpa.NoMessages().AddMany(msgs[:shareReqNNodes])
	// save the rest of the messages to be sent later
	msgs = msgs[shareReqNNodes:]
	if len(msgs) > 0 {
		m.shareReqMsgsMutex.Lock()
		m.shareReqMsgs[req.ID()] = &scheduledMessages{
			messages: msgs,
			lastSent: time.Now(),
		}
		m.shareReqMsgsMutex.Unlock()
	}
	return ret
}

// Message handles INCOMMING messages (from other nodes)
func (m *Impl) Message(msg gpa.Message) gpa.OutMessages {
	switch message := msg.(type) {
	case *msgShareRequest:
		return m.receiveMsgShareRequests(message)
	case *msgMissingRequest:
		return m.receiveMsgMissingRequest(message)
	default:
		m.log.Warnf("unexpected message %T: %+v", msg, msg)
	}
	return nil
}

func (m *Impl) receiveMsgShareRequests(msg *msgShareRequest) gpa.OutMessages {
	res := m.receiveRequests(msg.req)
	if len(res) != 1 || !res[0] {
		// message was rejected by the mempool
		return nil
	}
	if !msg.shouldPropagate {
		// reponses to "missing requests" are sent with shouldPropagate = false
		m.missingReqMsgsMutex.Lock()
		defer m.missingReqMsgsMutex.Unlock()
		reqid := msg.req.ID()
		delete(m.missingReqMsgs, reqid) // request received, no need to send any more messages
		return nil
	}
	return m.newShareRequestMessages(msg.req, msg.Sender())
}

// sender of the message is missing a request
func (m *Impl) receiveMsgMissingRequest(input *msgMissingRequest) gpa.OutMessages {
	req := m.getRequest(input.ref.ID)
	if req == nil {
		return nil
	}
	if !input.ref.IsFor(req) {
		m.log.Warnf("mismatch between requested requestRef and request in mempool. refHash: %s request:%s", input.ref.Hash.Hex(), req.String())
		return nil
	}
	return gpa.NoMessages().Add(newMsgShareRequest(req, false, input.Sender()))
}

func (m *Impl) UnmarshalMessage(data []byte) (msg gpa.Message, err error) {
	switch data[0] {
	case msgTypeMissingRequest:
		msg = &msgMissingRequest{}
	case msgTypeShareRequest:
		msg = &msgShareRequest{}
	default:
		return nil, fmt.Errorf("unknown message type %b", data[0])
	}
	err = msg.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (*Impl) Output() gpa.Output {
	// not used, callbacks are called instead
	return nil
}

// StatusString implements gpa.GPA
func (*Impl) StatusString() string {
	return "unimplemented"
}
