package main

import (
	"context"
	"fmt"
	"log"
)

type instruction int

const (
	Ping instruction = iota
	Heartbeat
	UpdateHeading
	UpdateStatus
	RebootSensors
	RebootConnection
	RestartEngine
	CloseConsole
	InstructionRange = 8
)

func main() {
	spacecraft := newSpacecraft()
	err := spacecraft.initialize()
	if err != nil {
		log.Fatalf("could not initialize spacecraft", "error", err)
		return
	}
	spacecraft.engine.start(spacecraft)
	spacecraft.engine.stop(spacecraft)
}

type console struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *console) start(s *spacecraft) {
	s.initwg.Done()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	for {
		select {
		case <-c.ctx.Done():
			fmt.Printf("%s: closing console", s.info.name)
			return
		default:
			var input int
			fmt.Printf("%s>", s.info.name)
			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Println("Error reading input:", err)
				continue
			}

			if input < 1 || input >= InstructionRange {
				fmt.Println("Invalid instruction. Enter a valid integer.")
				continue
			}

			err = c.handler(instruction(s, input))
			if err != nil {
				fmt.Printf("failed to execute console instruction: %w\n", err)
				continue
			}
		}
	}
}

func (c *console) handler(s *spacecraft, instr instruction) error {
	switch instr {
	case Ping:
		err := s.remoteBridge.ping()
		if err != nil {
			return fmt.Errorf("error sending message through remotebridge: %w", err)
		}
		return nil
	case CloseConsole:
		c.cancel()
	}
	return fmt.Errorf("failed to recognize instruction: %d", instr)
}
