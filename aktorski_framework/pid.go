package aktorski_framework

type PID struct {
	ID      string
	Address string
	Inbox   chan Envelope
}

func (p *PID) String() string {
	return p.ID
}
