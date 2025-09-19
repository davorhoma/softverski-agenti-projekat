package main

import (
	"fmt"
	akt "softverski-agenti-projekat/aktorski_framework"
)

type AggregatorActor struct {
	bolnicaID        string
	numHospitals     int
	receivedUpdates  map[string][]float64
	trainingActorPID *akt.PID
}

func NewAggregatorActor(bolnicaID string, numHospitals int, trainingActorPID *akt.PID) *AggregatorActor {
	return &AggregatorActor{
		bolnicaID:        bolnicaID,
		numHospitals:     numHospitals,
		receivedUpdates:  make(map[string][]float64),
		trainingActorPID: trainingActorPID,
	}
}

func (a *AggregatorActor) Receive(ctx *akt.Context, msg akt.Message) {
	switch m := msg.(type) {
	case *akt.ModelUpdateMessage:
		if m.SenderID != a.bolnicaID {
			a.receivedUpdates[m.SenderID] = m.Params
			fmt.Printf("[%s AggregatorActor] Received update from %s.\n", a.bolnicaID, m.SenderID)
		} else {
			a.receivedUpdates[a.bolnicaID] = m.Params
			fmt.Printf("[%s AggregatorActor] Received local update from %s.\n", a.bolnicaID, m.SenderID)
		}

		if len(a.receivedUpdates) == a.numHospitals {
			fmt.Printf("[%s AggregatorActor] Aggregating models and sending back.\n", a.bolnicaID)

			aggregatedParams := a.aggregateModels()

			ctx.Send(a.trainingActorPID, akt.ModelUpdateMessage{
				SenderID: a.bolnicaID,
				Params:   aggregatedParams,
			})

			a.receivedUpdates = make(map[string][]float64)
		}
	default:
		fmt.Printf("msg.(type): %T\n", m)
	}
}

func (a *AggregatorActor) aggregateModels() []float64 {
	if len(a.receivedUpdates) == 0 {
		return nil
	}

	numParams := len(a.receivedUpdates[a.bolnicaID])

	aggregatedParams := make([]float64, numParams)

	for _, params := range a.receivedUpdates {
		if len(params) == numParams {
			for i := 0; i < numParams; i++ {
				aggregatedParams[i] += params[i]
			}
		}
	}

	numModels := float64(len(a.receivedUpdates))
	for i := 0; i < numParams; i++ {
		aggregatedParams[i] /= numModels
	}

	return aggregatedParams
}
