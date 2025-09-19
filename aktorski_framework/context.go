package aktorski_framework

import "fmt"

type Context struct {
	self   *PID
	sender *PID
	System *ActorSystem
	actor  Actor
}

func (c *Context) Self() *PID {
	return c.self
}

func (c *Context) Sender() *PID {
	return c.sender
}

func (c *Context) Parent() *PID {
	return c.self.Parent
}

func (c *Context) Send(to *PID, msg Message) {
	if to == nil {
		fmt.Println("[WARN] Send: target PID is nil")
		return
	}

	envelope := Envelope{
		Msg:    msg,
		Sender: nil,
	}

	fmt.Printf("to: address %s, ID %s\n", to.Address, to.ID)
	if to.Address != c.System.localAddress && to.Address != "" {
		err := c.System.remoteSend(to, envelope)
		if err != nil {
			fmt.Printf("[ERROR] remote send failed: %v\n", err)
		}
	} else {
		select {
		case to.Inbox <- envelope:
		default:
			fmt.Println("[WARN] Send: inbox full for", to.ID)
		}
	}
}

func (c *Context) Request(to *PID, msg Message) {
	if to == nil {
		fmt.Println("[WARN] Send: target PID is nil")
		return
	}

	senderPID := &PID{
		ID:      c.self.ID,
		Address: c.System.localAddress,
	}

	envelope := Envelope{
		Msg:    msg,
		Sender: senderPID,
	}

	if to.Address != c.System.localAddress && to.Address != "" {
		err := c.System.remoteSend(to, envelope)
		if err != nil {
			fmt.Printf("[ERROR] remote request failed: %v\n", err)
		}
	} else {
		select {
		case to.Inbox <- envelope:
		default:
			fmt.Println("[WARN] Request: inbox full for", to.ID)
		}
	}
}

func (c *Context) Reply(msg Message) {
	if c.Sender() != nil {
		c.Send(c.Sender(), msg)
	}
}
