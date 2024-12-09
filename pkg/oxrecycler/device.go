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
	msgch             chan json.RawMessage
	connCtx           context.Context
	cancelCtx         context.CancelFunc
	mu                sync.RWMutex
	wg                sync.WaitGroup
}

type MessageHead struct {
	ConnectionID  string      `json:"connection_id"`
	TransactionID string      `json:"transaction_id"`
	From          string      `json:"from"`
	Type          MessageType `json:"message_type"`
	MessageLength int         `json:"message_length"`
	Timestamp     time.Time   `json:"timestamp"`
}

type Message struct {
	Headers MessageHead    `json:"headers"`
	Payload MessagePayload `json:"payload"`
	Errors  []Error        `json:"errors,omitempty"`
}

type Error struct {
	ErrCode    string    `json:"err_code"`
	ErrMsg     string    `json:"err_msg"`
	AlertLevel string    `json:"alert_level"`
	Timestamp  time.Time `json:"timestamp"`
}

type MessagePayload json.RawMessage

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

func CreateDevice(config Config) *Device {
	return &Device{
		DeviceID:          config.Preset.DeviceID,
		Temperature:       config.Preset.Temperature,
		Mode:              config.Preset.Mode,
		LastMaintenance:   config.Preset.LastMaintenance,
		HeartbeatInterval: config.Preset.HeartbeatInterval,
		Errors:            config.Preset.Errors,
	}
}

func (d *Device) Start(config Config) {
	d.wg.Add(1) //d.wg.Add(3) go d.StartSimulation() go d.StartHeartbeat()
	go func() {
		defer d.wg.Done()
		d.Connect(config)
		d.Disconnect()
	}()
	d.wg.Wait()
	d.Shutdown()
}

func (d *Device) Connect(config Config) {
	err := d.InitializeConnection(config.Preset.TCPServerAddress)
	if err != nil {

		return
	}
	d.Connection.wg.Add(1)
	go d.readLoop()
	d.Connection.wg.Wait()
}

func (d *Device) Disconnect() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Connection.TCPConnection != nil {
		fmt.Printf("Device %s: Closing old connection to %s...\n", d.DeviceID, d.Connection.RemoteAddress)

		if d.Connection.cancelCtx != nil {
			d.Connection.cancelCtx()
		}
		if err := d.Connection.TCPConnection.Close(); err != nil {
			fmt.Printf("Device %s: Error closing connection: %v\n", d.DeviceID, err)
		}
		close(d.Connection.msgch)
		d.Connection.TCPConnection = nil
		d.Connection.Status = StatusDisconnected
		fmt.Printf("Device %s: Connection closed successfully.\n", d.DeviceID)
	}
}

func (d *Device) InitializeConnection(addr string) error {
	connCtx, cancelCtx := context.WithCancel(context.Background())
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	d.Connection = Connection{
		Status:            StatusDisconnected,
		RemoteAddress:     conn.RemoteAddr().String(),
		TCPConnection:     conn,
		LastCommunication: time.Now(),
		msgch:             make(chan json.RawMessage, 10),
		connCtx:           connCtx,
		cancelCtx:         cancelCtx,
	}

	d.Connection.Status = StatusInit
	err = d.Connection.performHandshake()
	if err != nil {
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
	defer func() {
		d.Disconnect()
	}()

	buf := make([]byte, 2048)
	for {
		select {
		case <-d.Connection.connCtx.Done():
			fmt.Println("ReadLoop: Connection context canceled")
			return
		default:
			n, err := d.Connection.TCPConnection.Read(buf)
			if err != nil {
				fmt.Printf("ReadLoop: Read error: %v\n", err)
				return
			}
			fmt.Printf("ReadLoop: Failed to unmarshal message: %v\n", string(buf[:n]))

			/*
				n, err := d.Connection.TCPConn.Read(buf)
				if err != nil {
					if err == net.ErrClosed || err.Error() == "use of closed network connection" {
						d.disconnect()
						fmt.Println("ReadLoop: Connection closed")
						return
					}
					fmt.Printf("ReadLoop: Read error: %v\n", err)
					return
				}
				if n > 0 {
					msg, err := unMarshalMessage(buf[:n])
					if err != nil {
						fmt.Printf("ReadLoop: Failed to unmarshal message: %v\n", err)
						continue
					}
					d.Connection.MsgCh <- *msg
				} else {
					fmt.Println("ReadLoop: Cannot read empty buffer slice")
					continue
				}
			*/
		}
	}
}

/*
func (d *Device) startHeartbeat() {
	defer d.wg.Done()
	hbInt, err := time.ParseDuration(d.HeartbeatInterval)
	if err != nil {
		hbInt = 5 * time.Second
	}

	ticker := time.NewTicker(hbInt)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if d.Connection.ConnStatus == "CONNECTED" {
				d.mu.Lock()
				heartbeatMessage := d.createHeartBeatMessage()
				d.mu.Unlock()
				d.sendHeartbeat(heartbeatMessage)
			}
		case <-d.Connection.ConnCtx.Done():
			fmt.Printf("Device %s stopping heartbeats\n", d.ID)
			return
		}
	}
}

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

func (d *Device) createHeartBeatMessage() Message[Heartbeat] {
	heartbeatMessage := Message[Heartbeat]{
		ID:        uuid.New().String(),
		From:      d.Connection.TCPConn.LocalAddr().String(),
		MsgType:   "HEARTBEAT",
		Payload:   d.createHeartbeatPayload(),
		Timestamp: time.Now(),
	}
	return heartbeatMessage
}

func (d *Device) createHeartbeatPayload() Heartbeat {
	heartbeatPayload := Heartbeat{
		ID:              d.ID,
		Status:          d.Status,
		Mode:            d.Mode,
		Temperature:     d.Temperature,
		LastMaintenance: d.LastMaintenance,
		Errors:          d.Errors,
	}
	return heartbeatPayload
}

func (d *Device) sendHeartbeat(heartbeatMessage Message[Heartbeat]) {
	messageJSON, err := marshalMessage(heartbeatMessage)
	if err != nil {
		fmt.Printf("Error marshalling message: %v\n", err)
		return
	}
	fmt.Println(string(messageJSON))
	d.Connection.TCPConn.Write(messageJSON)
	d.Connection.LastCommunication = time.Now()
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

func unMarshalMessage(messageJSON []byte) (*Message[any], error) {
	var message Message[any]
	err := json.Unmarshal(messageJSON, &message)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling command: %v", err)
	}
	switch message.MsgType {
	case "HEARTBEAT":
		var payload Heartbeat
		err := json.Unmarshal(messageJSON, &payload)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling heartbeat payload: %v", err)
		}
		message.Payload = payload
	case "INSTRUCTION":
		var payload Instruction
		err := json.Unmarshal(messageJSON, &payload)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling instruction payload: %v", err)
		}
		message.Payload = payload
	default:
		return nil, fmt.Errorf("unknown message type: %s", message.MsgType)
	}

	return &message, nil
}

func marshalMessage[T any](message Message[T]) ([]byte, error) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling payload: %v", err)
	}
	return messageJSON, nil
}
*/
