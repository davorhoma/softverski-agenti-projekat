package aktorski_framework

type Message interface{}

type Actor interface {
	Receive(ctx *Context, msg Message)
}

type Behaviour func(ctx *Context, msg Message)

type Envelope struct {
	Msg    Message
	Sender *PID
}
