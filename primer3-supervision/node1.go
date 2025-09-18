package main

import (
	"fmt"
	akt "softverski-agenti-projekat/aktorski_framework"
	"time"
)

type GreetMessage struct {
	Text string
}

type FailingActor struct{}

func (a *FailingActor) Receive(ctx *akt.Context, msg akt.Message) {
	switch m := msg.(type) {
	case string:
		fmt.Printf("FailingActor received a message: %s\n", m)
		time.Sleep(3 * time.Second)
		panic("I crashed!")
	}
}

func (a *FailingActor) Starting() {
	fmt.Println("Actor starting")
}

func main() {
	system1 := akt.NewActorSystem("localhost:8080")
	go akt.StartRemoteServer(system1, "localhost:8080")

	time.Sleep(1 * time.Second)

	supervisorPID := system1.Spawn("supervisor", nil, &akt.SupervisorActor{})
	failingActorPID := system1.Spawn("failing-actor", supervisorPID, &FailingActor{})

	fmt.Println("Node 1 is running. Sending a message to the failing actor.")

	system1.Send(failingActorPID.ID, "start")

	fmt.Println("Press Enter to exit.")
	fmt.Scanln()
}
