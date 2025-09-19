package main

import (
	"encoding/json"
	"fmt"
	akt "softverski-agenti-projekat/aktorski_framework"
	"time"
)

type IncrementMessage struct{}

type CounterState struct {
	Count int
}

type CounterActor struct {
	state          CounterState
	persistencePID *akt.PID
}

func NewCounterActor(persistencePID *akt.PID) *CounterActor {
	return &CounterActor{
		state:          CounterState{Count: 0},
		persistencePID: persistencePID,
	}
}

func (a *CounterActor) Recover(ctx *akt.Context) {
	fmt.Println("[CounterActor] Starting up and trying to recover state...")

	replyChan := make(chan []byte)

	request := akt.LoadStateRequest{
		Key:       "counter_state",
		ReplyChan: replyChan,
	}
	ctx.Send(a.persistencePID, request)

	select {
	case stateBytes := <-replyChan:
		if stateBytes != nil {
			var loadedState CounterState
			if err := json.Unmarshal(stateBytes, &loadedState); err != nil {
				fmt.Printf("[CounterActor] Failed to unmarshal state: %v\n", err)
				return
			}
			a.state = loadedState
			fmt.Printf("[CounterActor] State recovered. Current count is: %d\n", a.state.Count)
		}
	case <-time.After(5 * time.Second):
		fmt.Println("[CounterActor] Recovery timeout.")
	}
}

func (a *CounterActor) Receive(ctx *akt.Context, msg akt.Message) {
	switch m := msg.(type) {
	case IncrementMessage:
		a.state.Count++
		fmt.Printf("[CounterActor] Counter incremented to: %d\n", a.state.Count)

		saveMsg := akt.SaveStateMessage{
			Key:   "counter_state",
			State: a.state,
		}
		ctx.Send(a.persistencePID, saveMsg)

	case akt.StateLoadedMessage:
		var loadedState CounterState
		if err := json.Unmarshal(m.State, &loadedState); err != nil {
			fmt.Printf("[CounterActor] Failed to unmarshal state: %v\n", err)
			return
		}
		a.state = loadedState
		fmt.Printf("[CounterActor] State recovered. Current count is: %d\n", a.state.Count)

	default:
		fmt.Printf("[CounterActor] Received unknown message: %+v\n", msg)
	}
}
