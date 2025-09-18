package aktorski_framework

import "fmt"

type SupervisorActor struct{}

func (a *SupervisorActor) Receive(ctx *Context, msg Message) {
	switch m := msg.(type) {
	case FailureMessage:
		fmt.Printf("Supervisor received failure from %s: %s\n", m.FailedPID.ID, m.Reason)
		ctx.System.Stop(m.FailedPID)

		originalActorType, ok := ctx.System.actorTypes[m.FailedPID.ID]
		if !ok {
			fmt.Printf("[SupervisorActor] Cannot restart actor %s: actor type not found.\n", m.FailedPID.ID)
			return
		}

		fmt.Printf("[SupervisorActor] Restarting actor %s\n", m.FailedPID.ID)
		ctx.System.Spawn(m.FailedPID.ID, ctx.Self(), originalActorType)
	default:
		fmt.Printf("[SupervisorActor] Received a regular message: %+v\n", msg)
	}
}
