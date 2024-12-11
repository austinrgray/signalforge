package oxrecycler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Device struct {
	Config            Config
	DeviceID          string        `json:"device_id"`
	Temperature       float32       `json:"temperature"`
	Mode              DeviceMode    `json:"mode"`
	LastMaintenance   time.Time     `json:"last_maintenance"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	Errors            []Error       `json:"errors"`
	Connection        Connection
	mu                sync.RWMutex
	wg                sync.WaitGroup
}

type Connection struct {
	ConnectionID      string
	Status            ConnectionStatus
	RemoteAddress     string
	TCPConnection     net.Conn
	LastCommunication time.Time
	msgch             chan Message[any]
	connCtx           context.Context
	mu                sync.RWMutex
	wg                sync.WaitGroup
}

type Message[T any] struct {
	Headers MessageHead `json:"headers"`
	Payload T           `json:"payload"`
	Errors  []Error     `json:"errors,omitempty"`
}

type MessageHead struct {
	ConnectionID  string      `json:"connection_id"`
	TransactionID string      `json:"transaction_id"`
	From          string      `json:"from"`
	Type          MessageType `json:"message_type"`
	Timestamp     time.Time   `json:"timestamp"`
}

type HeartbeatPayload struct {
	DeviceID        string           `json:"device_id"`
	Temperature     float32          `json:"temperature"`
	Mode            DeviceMode       `json:"mode"`
	LastMaintenance time.Time        `json:"last_maintenance"`
	Status          ConnectionStatus `json:"connection_status"`
}

type Error struct {
	Code        string    `json:"error_code"`
	Description string    `json:"error_message"`
	AlertLevel  string    `json:"alert_level"`
	Timestamp   time.Time `json:"timestamp"`
}

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "CONNECTED"
	StatusDisconnected ConnectionStatus = "DISCONNECTED"
	StatusReconnected  ConnectionStatus = "RECONNECTING"
	StatusInit         ConnectionStatus = "INITIALIZING"
	StatusAuth         ConnectionStatus = "AUTHENTICATING"
	StatusLockout      ConnectionStatus = "LOCKED OUT"
)

type DeviceMode string

const (
	ModeLowPower      DeviceMode = "LOW_POWER_USEAGE"
	ModeAutoAdjust    DeviceMode = "AUTO"
	ModeRapidRecovery DeviceMode = "RAPID_RECOVERY"
	ModeActive        DeviceMode = "NORMAL"
	ModeIdle          DeviceMode = "IDLE"
	ModeMaintenance   DeviceMode = "MAINTENANCE"
	ModeError         DeviceMode = "ERROR"
	ModeReboot        DeviceMode = "REBOOT"
	ModeSleep         DeviceMode = "SLEEP"
	ModeOff           DeviceMode = "OFF"
)

type MessageType string

const (
	MsgTypeHandshake    MessageType = "handshake"
	MsgTypeData         MessageType = "data"
	MsgTypeError        MessageType = "error"
	MsgTypeHeartbeat    MessageType = "heartbeat"
	MsgTypeAck          MessageType = "ack"
	MsgTypeAuthRequest  MessageType = "auth_request"
	MsgTypeAuthResponse MessageType = "auth_response"
	MsgTypeCommand      MessageType = "command"
	MsgTypeResponse     MessageType = "response"
)

func (d *Device) Start() {
	d.wg.Add(1)
	//go d.StartSimulation()
	go d.Connect()

	d.wg.Wait()
	d.Shutdown()
}

func (d *Device) Connect() {
	defer d.wg.Done()
	d.wg.Add(1)
	err := d.InitializeConnection()
	if err != nil {
		fmt.Println("error initializing connection")
		return
	}
	select {
	case <-d.Connection.connCtx.Done():
		fmt.Println("Connect: connection context has ended")
		return
	default:
		d.Connection.wg.Add(2)
		go d.readLoop()
		go d.startHeartbeat()
		d.Connection.wg.Wait()
	}
}

func (d *Device) Disconnect() {
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.wg.Done()

	if d.Connection.TCPConnection != nil {
		fmt.Printf("Device %s: Closing old connection to %s...\n", d.DeviceID, d.Connection.RemoteAddress)

		if err := d.Connection.TCPConnection.Close(); err != nil {
			fmt.Printf("Device %s: Error closing connection: %v\n", d.DeviceID, err)
		}
		close(d.Connection.msgch)
		d.Connection.TCPConnection = nil
		d.Connection.Status = StatusDisconnected
		fmt.Printf("Device %s: Connection closed successfully.\n", d.DeviceID)
	}
}

func (d *Device) InitializeConnection() error {
	defer d.wg.Done()
	connCtx := context.Background()
	conn, err := net.Dial("tcp", d.Config.TCPServerHost+d.Config.TCPServerPort)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	d.Connection = Connection{
		Status:            StatusDisconnected,
		RemoteAddress:     conn.RemoteAddr().String(),
		TCPConnection:     conn,
		LastCommunication: time.Now(),
		msgch:             make(chan Message[any], 10),
		connCtx:           connCtx,
	}
	d.Connection.mu.Lock()
	d.Connection.Status = StatusInit
	d.Connection.mu.Unlock()
	err = d.Connection.performHandshake()
	if err != nil {
		d.wg.Add(1)
		d.Disconnect()
		return fmt.Errorf("handshake failed: %w", err)
	}

	d.Connection.Status = StatusConnected
	log.Printf("Connection initialized with ID: %s\n", d.Connection.ConnectionID)
	return nil

}

func (c *Connection) performHandshake() error {
	ctx, cancel := context.WithTimeout(c.connCtx, 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("handshake timed out")
	default:

		serverPublicKey, err := c.receiveServerPublicKey()
		if err != nil {
			return fmt.Errorf("failed to receive server public key: %w", err)
		}
		fmt.Printf("Received server public key: %s\n", serverPublicKey)

		clientToken := generateClientToken()

		err = c.sendClientToken(clientToken)
		if err != nil {
			return fmt.Errorf("failed to send client token: %w", err)
		}
		fmt.Println("Client token sent successfully")

		err = c.receiveConnectionID()
		if err != nil {
			return fmt.Errorf("failed to receive connectionID: %w", err)
		}
		fmt.Printf("Received connectionID: %s\n", c.ConnectionID)
	}

	return nil
}

func (c *Connection) receiveServerPublicKey() (string, error) {
	buffer := make([]byte, 4096)
	n, err := c.TCPConnection.Read(buffer)
	if err != nil {
		return "", err
	}
	return string(buffer[:n]), nil
}

func (c *Connection) receiveConnectionID() error {
	buffer := make([]byte, 4096)
	n, err := c.TCPConnection.Read(buffer)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.ConnectionID = string(buffer[:n])
	c.mu.Unlock()

	return nil
}

func generateClientToken() string {
	return "valid-client-token"
}

func (c *Connection) sendClientToken(token string) error {
	_, err := c.TCPConnection.Write([]byte(token))
	return err
}

func (d *Device) Shutdown() {
	d.Disconnect()
	d.wg.Wait()
	fmt.Printf("Device %s has shut down\n", d.DeviceID)
}

func (d *Device) readLoop() {
	defer d.Connection.wg.Done()
	defer d.Connection.TCPConnection.Close()

	buf := make([]byte, 2048)
	for {
		select {
		case <-d.Connection.connCtx.Done():
			fmt.Println("ReadLoop: Connection context canceled")
			return
		default:
			n, err := d.Connection.TCPConnection.Read(buf)
			if err != nil {
				if err == net.ErrClosed || err.Error() == "use of closed network connection" {
					fmt.Println("ReadLoop: Connection closed")
					return
				}
				fmt.Printf("ReadLoop: Read error: %v\n", err)
				return
			}
			if n > 0 {
				msg, err := UnMarshalMessage(buf[:n])
				if err != nil {
					fmt.Printf("ReadLoop: Failed to unmarshal message: %v\n", err)
					continue
				}
				d.Connection.msgch <- *msg
			}
		}
	}
}

func (d *Device) startHeartbeat() {
	defer d.Connection.wg.Done()
	hbInt := d.HeartbeatInterval
	ticker := time.NewTicker(hbInt)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if d.Connection.Status == StatusConnected {
				d.mu.RLock()
				heartbeatMessage := d.createHeartBeatMessage()
				d.mu.RUnlock()

				d.sendHeartbeat(*heartbeatMessage)
			}
		case <-d.Connection.connCtx.Done():
			fmt.Printf("Device %s stopping heartbeats\n", d.DeviceID)
			return
		}

	}
}

func (d *Device) createMessageHead() *MessageHead {
	return &MessageHead{
		ConnectionID:  d.Connection.ConnectionID,
		TransactionID: "test-123", // Placeholder for transaction ID; adjust as necessary
		From:          d.Connection.TCPConnection.LocalAddr().String(),
		Type:          MsgTypeHeartbeat,
		Timestamp:     time.Now().UTC(),
	}
}

func (d *Device) createHeartBeatMessage() *Message[HeartbeatPayload] {
	payload := HeartbeatPayload{
		DeviceID:        d.DeviceID,
		Temperature:     d.Temperature,
		Mode:            d.Mode,
		LastMaintenance: d.LastMaintenance,
		Status:          d.Connection.Status,
	}

	return &Message[HeartbeatPayload]{
		Headers: *d.createMessageHead(),
		Payload: payload,
		Errors:  d.Errors,
	}
}

func (d *Device) sendHeartbeat(message Message[HeartbeatPayload]) {
	messageJSON, err := message.MarshalMessage()
	if err != nil {
		fmt.Printf("Error marshalling message: %v\n", err)
		return
	}
	message.PrintMessage()
	d.Connection.TCPConnection.Write(messageJSON)
	d.mu.Lock()
	d.Connection.LastCommunication = time.Now()
	d.mu.Unlock()
}

func (m *Message[T]) MarshalMessage() ([]byte, error) {
	messageJSON, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshalling full message: %v", err)
	}
	return messageJSON, nil
}

func UnMarshalMessage(raw []byte) (*Message[any], error) {
	var message Message[any]
	err := json.Unmarshal(raw, &message)
	if err != nil {
		return &message, fmt.Errorf("error marshalling full message: %v", err)
	}
	switch message.Headers.Type {
	case MsgTypeHeartbeat:
		var payload HeartbeatPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return &message, fmt.Errorf("error unmarshalling heartbeat payload: %v", err)
		}
		message.Payload = payload
	default:
		return &message, fmt.Errorf("unknown message type: %v", message.Headers.Type)
	}
	return &message, nil
}

func (m *Message[T]) PrintMessage() {
	msgJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}
	fmt.Println(string(msgJSON))
}

/*


func (d *Device) handleMessage(msg Message[any]) {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch msg.MsgType {
	case "INSTRUCTION":
		d.processInstruction(instructionMessage)
	case "PING":
		d.mu.Lock()
		heartbeatMessage := d.createHeartBeatMessage()
		d.mu.Unlock()
		d.sendHeartbeat(heartbeatMessage)
	}
}

func (d *Device) processInstruction(instruction Instruction) {
	fmt.Printf("Device %s received command: %s\n", d.ID, string(instruction.Instruction))

	switch instruction.Instruction {
	case "STATUS_CHANGE":
		d.Status = instruction.Argument
		fmt.Printf("Device %s updated status to: %s\n", d.ID, d.Status)
	case "REBOOT":
		d.reboot()
	case "RECONNECT":
		d.reconnect()
	default:
		fmt.Printf("Device %s received unknown command: %s\n", d.ID, instruction.Instruction)
	}
}

func (d *Device) reboot() {
	fmt.Printf("Device %s is rebooting...\n", d.ID)
	time.Sleep(15 * time.Second) // Simulate reboot delay.
	fmt.Printf("Device %s has rebooted\n", d.ID)
}



func (d *Device) reconnect() {
	d.mu.Lock()
	d.wg.Wait()
	if d.Connection.TCPConn != nil {
		fmt.Printf("Device %s: Closing old connection to %s...\n", d.ID, d.Connection.TCPAddr)

		if d.Connection.CancelCtx != nil {
			d.Connection.CancelCtx()
		}
		if err := d.Connection.TCPConn.Close(); err != nil {
			fmt.Printf("Device %s: Error closing connection: %v\n", d.ID, err)
		}
		close(d.Connection.MsgCh)

		fmt.Printf("Device %s: Connection closed successfully.\n", d.ID)
		d.Connection.ConnStatus = "RECONNECTING"
	}

	d.mu.Unlock()

	fmt.Printf("Device %s is reconnecting...\n", d.ID)
	if err := d.Connect(); err != nil {
		fmt.Printf("Device %s: Reconnect failed: %v\n", d.ID, err)
		return
	}
}

*/
