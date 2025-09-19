package main

import (
	"fmt"
	"time"

	akt "softverski-agenti-projekat/aktorski_framework"
)

func main() {
	system := akt.NewActorSystem("localhost:8080")

	persistenceActor, err := akt.NewPersistenceActor("actor_states.db")
	if err != nil {
		panic(err)
	}
	persistencePID := system.Spawn("persistence", nil, persistenceActor)

	time.Sleep(5 * time.Second)

	counterPID := system.Spawn("counter", nil, NewCounterActor(persistencePID))

	time.Sleep(1 * time.Second)

	fmt.Println("Sending 3 messages to increment the counter.")
	for i := 0; i < 3; i++ {
		system.Send(counterPID.ID, IncrementMessage{})
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\nStopping the system and restarting to check persistence.")
	system.Stop(counterPID)
	system.Stop(persistencePID)
	system = nil

	fmt.Println("Restarting the system...")
	system = akt.NewActorSystem("localhost:8080")
	fmt.Println("Created new actor system")
	persistenceActor, err = akt.NewPersistenceActor("actor_states.db")
	if err != nil {
		panic(err)
	}
	fmt.Println("Created persistenceActor")

	persistencePID = system.Spawn("persistence", nil, persistenceActor)
	counterPID = system.Spawn("counter", nil, NewCounterActor(persistencePID))
	
	time.Sleep(1 * time.Second)
	
	fmt.Println("\nSending another increment message. The count should continue.")
	system.Send(counterPID.ID, IncrementMessage{})
	
	fmt.Println("Press Enter to exit.")
	fmt.Scanln()
}