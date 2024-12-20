package main

import (
	"context"
	"fmt"
	"net"
	"sync"
)

type remoteBridge struct {
	sessionID      string
	remoteHostAddr string
	rch            chan []byte
	wch            chan message
	conn           net.Conn
	hbInt          uint
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	lock           sync.RWMutex
}

func (r *remoteBridge) initialize(s *spacecraft) error {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	conn, err := net.Dial("tcp", r.remoteHostAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w\n", err)
	}

	r.conn = conn

	r.sessionID, err = r.handshake(s)
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	return nil
}

func (r *remoteBridge) connect(s *spacecraft) {
	r.wg.Add(3)
	go r.reader()
	go r.writer()
	go r.handler()
	r.wg.Wait()
	r.disconnect()
}

func (r *remoteBridge) disconnect() {
	r.lock.Lock()
	r.cancel()

	if r.conn != nil {
		_, err := r.conn.Write([]byte("disconnecting from remote bridge"))
		if err != nil {
			fmt.Printf("could not send disconnect: %w\n", err)
		}
		r.conn.Close()
		r.conn = nil
	}
	if r.rch != nil {
		close(r.rch)
	}
	if r.wch != nil {
		close(r.wch)
	}
	r.lock.Unlock()
	fmt.Println("disconnect: successfully disconnected from remote bridge")
}

func (r *remoteBridge) handshake(s *spacecraft) (string, error) {
	clientHello := s.info.id + ":" + " requesting session id"

	_, err := r.conn.Write([]byte(clientHello))
	if err != nil {
		return "", fmt.Errorf("failed to write: %w", err)
	}

	buf := make([]byte, 256)
	n, err := r.conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read: %w", err)
	}

	return string(buf[:n]), nil
}

func (r *remoteBridge) reader() {
	defer r.wg.Done()

	buf := make([]byte, 2048)
	for {
		select {
		case <-r.ctx.Done():
			fmt.Println("reader: connection context is canceled")
			return
		default:
			n, err := r.conn.Read(buf)
			if err != nil {
				fmt.Println("reader: error reading from %s", r.remoteHostAddr)
			}
			r.rch <- buf[:n]
			fmt.Printf("reader: message received from %s", r.remoteHostAddr)
		}
	}
}

func (r *remoteBridge) writer() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			fmt.Println("writer: connection context is canceled")
			return
		case msg := <-r.wch:
			_ = msg
			_, err := r.conn.Write([]byte("test")) //need message struct/marshaling
			if err != nil {
				fmt.Printf("writer: error writing to %s", r.remoteHostAddr)
			}
			fmt.Printf("writer: message sent to %s", r.remoteHostAddr)
		}
	}
}

func (r *remoteBridge) handler() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			fmt.Println("connection handler: connection context is canceled")
			return
		case msg := <-r.rch:
			_ = msg

		}
	}
}

type envelope struct {
	version uint8
	sender string
	sessionID string
	traceID string
	length uint16
	oneof

	
}

type header struct {
	version uint8
	command uint8

}
type ping


func (r *remoteBridge) ping() {
	msg := &message{}
	r.wch <- msg
}
