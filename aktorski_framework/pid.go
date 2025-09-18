package aktorski_framework

type PID struct {
	ID      string
	Address string
	Inbox   chan Envelope
	Parent  *PID
}

func (p *PID) String() string {
	return p.ID
}
