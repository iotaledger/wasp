package ledger_go

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// SpeculosTransportOpts contains configuration for the transport
type SpeculosTransportOpts struct {
	ApduPort       int
	ButtonPort     *int
	AutomationPort *int
	Host           string
}

// SpeculosTransport implements a TCP transport for Speculos
type SpeculosTransport struct {
	apduSocket       net.Conn
	opts             SpeculosTransportOpts
	automationSocket net.Conn
	automationEvents chan map[string]interface{}
}

func DefaultSpeculosTransportOpts() SpeculosTransportOpts {
	return SpeculosTransportOpts{
		ApduPort: 40000,
		Host:     "127.0.0.1",
	}
}

// NewSpeculosTransport creates a new transport instance
func NewSpeculosTransport(opts SpeculosTransportOpts) (*SpeculosTransport, error) {
	// Connect to APDU socket
	apduSocket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.ApduPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to APDU socket: %w", err)
	}

	transport := &SpeculosTransport{
		apduSocket:       apduSocket,
		opts:             opts,
		automationEvents: make(chan map[string]interface{}, 100),
	}

	// Setup automation socket if port is specified
	if opts.AutomationPort != nil {
		autoSocket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, *opts.AutomationPort))
		if err != nil {
			apduSocket.Close()
			return nil, fmt.Errorf("failed to connect to automation socket: %w", err)
		}
		transport.automationSocket = autoSocket

		// Handle automation events in a goroutine
		go transport.handleAutomationEvents()
	}

	// Wait a bit to ensure connection is stable
	time.Sleep(100 * time.Millisecond)
	return transport, nil
}

// Button sends a button command to Speculos
func (t *SpeculosTransport) Button(command string) error {
	if t.opts.ButtonPort == nil {
		return fmt.Errorf("buttonPort is missing")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", t.opts.Host, *t.opts.ButtonPort))
	if err != nil {
		return fmt.Errorf("failed to connect to button port: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command))
	return err
}

// Exchange sends an APDU command and receives the response
func (t *SpeculosTransport) Exchange(apdu []byte) ([]byte, error) {

	// Encode and send APDU
	encoded := encodeAPDU(apdu)

	_, err := t.apduSocket.Write(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to write APDU: %w", err)
	}

	// Read response
	sizeBuf := make([]byte, 4)
	_, err = t.apduSocket.Read(sizeBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response size: %w", err)
	}

	//
	size := binary.BigEndian.Uint32(sizeBuf)
	response := make([]byte, size+2) // +2 for status code
	_, err = t.apduSocket.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	status := codec.Uint16(response[size:])
	if status != 0x9000 {
		return nil, fmt.Errorf("invalid status response: %x", status)
	}

	return response[:len(response)-2], nil
}

// Close closes all connections
func (t *SpeculosTransport) Close() error {
	if t.automationSocket != nil {
		t.automationSocket.Close()
	}
	return t.apduSocket.Close()
}

func (t *SpeculosTransport) handleAutomationEvents() {
	buf := make([]byte, 1024)
	for {
		n, err := t.automationSocket.Read(buf)
		if err != nil {
			// Handle error or connection close
			close(t.automationEvents)
			return
		}

		var event map[string]interface{}
		err = json.Unmarshal(buf[:n], &event)
		if err != nil {
			continue
		}

		t.automationEvents <- event
	}
}

func encodeAPDU(apdu []byte) []byte {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(apdu)))
	return append(size, apdu...)
}
