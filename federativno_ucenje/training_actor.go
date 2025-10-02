package main

import (
	"encoding/json"
	"fmt"
	akt "softverski-agenti-projekat/aktorski_framework"
	utils "softverski-agenti-projekat/federativno_ucenje/utils"
	"time"
)

type ModelParamsMessage struct {
	Params []float64
}

type TrainingActor struct {
	modelParams      []float64
	bolnicaID        string
	peers            []*akt.PID
	persistencePID   *akt.PID
	aggregatorPID    *akt.PID
	trainFilePath    string
	testFilePath     string
	currentBehaviour akt.Behaviour
}

func NewTrainingActor(bolnicaID string, persistencePID *akt.PID, peers []*akt.PID, aggregatorPID *akt.PID, trainFilePath, testFilePath string) *TrainingActor {
	a := &TrainingActor{
		bolnicaID:      bolnicaID,
		persistencePID: persistencePID,
		peers:          peers,
		aggregatorPID:  aggregatorPID,
		trainFilePath:  trainFilePath,
		testFilePath:   testFilePath,
	}

	a.currentBehaviour = a.trainingBehaviour
	return a
}

func (a *TrainingActor) become(b akt.Behaviour) {
	a.currentBehaviour = b
}

func (a *TrainingActor) Recover(ctx *akt.Context) {
	fmt.Printf("[%s TrainingActor] Starting up and trying to recover model from persistence...\n", a.bolnicaID)

	replyChan := make(chan []byte)
	request := akt.LoadStateRequest{
		Key:       fmt.Sprintf("%s_model_params", a.bolnicaID),
		ReplyChan: replyChan,
	}
	ctx.Send(a.persistencePID, request)

	select {
	case stateBytes := <-replyChan:
		if stateBytes != nil {
			var loadedParams []float64
			if err := json.Unmarshal(stateBytes, &loadedParams); err != nil {
				fmt.Printf("[%s TrainingActor] Failed to unmarshal model params: %v\n", a.bolnicaID, err)
				return
			}
			a.modelParams = make([]float64, len(loadedParams))
			copy(a.modelParams, loadedParams)
			fmt.Printf("[%s TrainingActor] Model parameters recovered.\n", a.bolnicaID)
		} else {
			fmt.Printf("[%s TrainingActor] No previous model found.", a.bolnicaID)
		}
	case <-time.After(5 * time.Second):
		fmt.Printf("[%s TrainingActor] Recovery timeout. Initializing a new model.\n", a.bolnicaID)
	}
}

func (a *TrainingActor) Receive(ctx *akt.Context, msg akt.Message) {
	a.currentBehaviour(ctx, msg)
}

func (a *TrainingActor) trainingBehaviour(ctx *akt.Context, msg akt.Message) {
	switch msg.(type) {
	case *akt.TrainModelMessage:
		fmt.Printf("[%s TrainingActor] Starting local training...\n", a.bolnicaID)

		updatedParams, err := utils.TrainHeartDiseaseLogisticRegression(a.trainFilePath, a.modelParams)
		if err != nil {
			fmt.Printf("[%s TrainingActor] Training failed: %v\n", a.bolnicaID, err)
			return
		}
		a.modelParams = make([]float64, len(updatedParams))
		copy(a.modelParams, updatedParams)
		fmt.Printf("[%s TrainingActor] Training finished. Sending updates to peers.\n", a.bolnicaID)

		// Slanje ažuriranih parametara drugim bolnicama
		for _, peer := range a.peers {
			ctx.Send(peer, &akt.ModelUpdateMessage{
				SenderID: a.bolnicaID,
				Params:   a.modelParams,
			})
		}

		// Slanje ažuriranih parametara lokalnom agregatoru
		ctx.Send(a.aggregatorPID, &akt.ModelUpdateMessage{
			SenderID: a.bolnicaID,
			Params: a.modelParams,
		})

		// Čuvanje stanja
		saveMsg := akt.SaveStateMessage{
			Key:   fmt.Sprintf("%s_model_params", a.bolnicaID),
			State: a.modelParams,
		}
		ctx.Send(a.persistencePID, saveMsg)

		// Evaluacija
		rez, _ := utils.EvaluateHeartDiseaseLogisticRegression(a.testFilePath, a.modelParams)
		fmt.Printf("[%s TrainingActor] Evaluation: %v\n", a.bolnicaID, rez)

		a.become(a.waitingForAggregationBehaviour)

		go func() {
			time.Sleep(15 * time.Second)
			ctx.Send(ctx.Self(), akt.TrainModelMessage{})
		}()

	default:
		fmt.Println("IGNORISANA PORUKA")
	}
}

func (a *TrainingActor) waitingForAggregationBehaviour(ctx *akt.Context, msg akt.Message) {
	switch m := msg.(type) {
	case akt.ModelUpdateMessage:
		fmt.Printf("[%s TrainingActor] Received aggregated model. Updating parameters.\n", a.bolnicaID)
		a.modelParams = make([]float64, len(m.Params))
		copy(a.modelParams, m.Params)

		rez, _ := utils.EvaluateHeartDiseaseLogisticRegression(a.testFilePath, a.modelParams)
		fmt.Printf("[%s TrainingActor] Evaluation with aggregated model: %v\n", a.bolnicaID, rez)

		a.become(a.trainingBehaviour)
		ctx.Send(ctx.Self(), &akt.TrainModelMessage{})
	case akt.TrainModelMessage, *akt.TrainModelMessage:
		fmt.Printf("[%s TrainingActor] Timeout waiting for aggregated model. Starting new round with local model.\n", a.bolnicaID)
		a.become(a.trainingBehaviour)
		ctx.Send(ctx.Self(), &akt.TrainModelMessage{})
	}
}
