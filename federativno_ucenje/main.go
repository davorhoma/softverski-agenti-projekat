package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	akt "softverski-agenti-projekat/aktorski_framework"
)

type HospitalConfig struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <hospital_id>")
		os.Exit(1)
	}
	myHospitalID := os.Args[1]

	configData, err := os.ReadFile("hospitals.json")
	if err != nil {
		panic(fmt.Errorf("failed to read hospitals.json: %w", err))
	}
	var hospitals []HospitalConfig
	if err := json.Unmarshal(configData, &hospitals); err != nil {
		panic(fmt.Errorf("failed to parse hospitals.json: %w", err))
	}

	var myAddress string
	var peerPIDs []*akt.PID

	for _, h := range hospitals {
		if h.ID == myHospitalID {
			myAddress = h.Address
		} else {
			peerPID := &akt.PID{
				ID:      fmt.Sprintf("%s-aggregator", h.ID),
				Address: h.Address,
			}
			peerPIDs = append(peerPIDs, peerPID)
		}
	}

	if myAddress == "" {
		panic("Hospital ID not found in config file")
	}

	dbPath := fmt.Sprintf("%s_actor_states.db", myHospitalID)
	trainFile := fmt.Sprintf("data/framingham-%s-train.csv", myHospitalID[len(myHospitalID)-1:])
	testFile := fmt.Sprintf("data/framingham-%s-test.csv", myHospitalID[len(myHospitalID)-1:])

	system := akt.NewActorSystem(myAddress)

	go akt.StartRemoteServer(system, myAddress)

	persistenceActor, err := akt.NewPersistenceActor(dbPath)
	if err != nil {
		panic(err)
	}
	persistencePID := system.Spawn("persistence", nil, persistenceActor)

	trainingActor := NewTrainingActor(myHospitalID, persistencePID, peerPIDs, nil, trainFile, testFile)
	trainingPID := system.Spawn(fmt.Sprintf("%s-training", myHospitalID), nil, trainingActor)

	aggregatorActor := NewAggregatorActor(myHospitalID, len(hospitals), trainingPID)
	aggregatorPID := system.Spawn(fmt.Sprintf("%s-aggregator", myHospitalID), nil, aggregatorActor)

	trainingActor.aggregatorPID = aggregatorPID

	time.Sleep(3 * time.Second)

	fmt.Printf("--- %s: Inicijalizacija prve runde treniranja. ---\n", myHospitalID)
	system.Send(trainingPID.ID, &akt.TrainModelMessage{})

	fmt.Println("\nGotovo. Pritisnite Enter za izlaz.")
	fmt.Scanln()
}
