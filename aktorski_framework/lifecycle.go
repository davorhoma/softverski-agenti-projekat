package aktorski_framework

type PreStart interface {
	Starting()
}

type PostStart interface {
	Started()
}

type PostStop interface {
	Stopped()
}

type Restart interface {
	Restart(err error)
}

type PreRecover interface {
	Recover(ctx *Context)
}
