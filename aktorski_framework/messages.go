package aktorski_framework

type FailureMessage struct {
	FailedPID *PID
	Reason    string
}

type SaveStateMessage struct {
	Key   string
	State interface{}
}

type LoadStateMessage struct {
	Key string
}

type LoadStateRequest struct {
	Key       string
	ReplyChan chan []byte
}

type StateLoadedMessage struct {
	Key   string
	State []byte
}
