package aktorski_framework

import (
	"context"
	"encoding/json"
	"fmt"
	pb "softverski-agenti-projekat/aktorski_framework/messages"
	"sync"

	"google.golang.org/grpc"
)

type remoteClient struct {
	client pb.ActorServiceClient
	conn   *grpc.ClientConn
}

type ActorSystem struct {
	actors        map[string]*PID
	actorTypes    map[string]Actor
	mu            sync.RWMutex
	remoteClients map[string]*remoteClient
	remoteMu      sync.RWMutex
	localAddress  string
}

func NewActorSystem(address string) *ActorSystem {
	return &ActorSystem{
		actors:        make(map[string]*PID),
		actorTypes:    make(map[string]Actor),
		localAddress:  address,
		remoteClients: make(map[string]*remoteClient),
	}
}

func (s *ActorSystem) RegisterActor(pid *PID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actors[pid.ID] = pid
}

func (s *ActorSystem) GetPID(id string) (*PID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pid, ok := s.actors[id]
	return pid, ok
}

func (s *ActorSystem) Spawn(id string, parent *PID, actor Actor) *PID {
	inbox := make(chan Envelope, 100)
	pid := &PID{
		ID:      id,
		Address: s.localAddress,
		Inbox:   inbox,
		Parent:  parent,
	}

	s.actorTypes[id] = actor

	s.RegisterActor(pid)

	if ps, ok := actor.(PreStart); ok {
		ps.Starting()
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Actor %s panicked: %v\n", pid.ID, r)
				// Obaveštavanje roditelja o panic-u
				if pid.Parent != nil {
					s.Send(pid.Parent.ID, FailureMessage{
						FailedPID: pid,
						Reason:    fmt.Sprintf("%v", r),
					})
				}
			}
		}()

		for env := range inbox {
			ctx := &Context{
				self:   pid,
				System: s,
				sender: env.Sender,
			}

			behavior := ctx.CurrentBehaviour()
			if behavior != nil {
				behavior(ctx, env.Msg)
			} else {
				actor.Receive(ctx, env.Msg)
			}
		}

		if ps, ok := actor.(PostStop); ok {
			ps.Stopped()
		}
	}()

	return pid
}

func (s *ActorSystem) Send(id string, msg Message) {
	s.mu.RLock()
	pid, ok := s.actors[id]
	s.mu.RUnlock()

	if !ok {
		fmt.Println("[WARN] Send: actor", id, "not found")
		return
	}

	envelope := Envelope{
		Msg:    msg,
		Sender: nil,
	}

	// Slanje poruke u sanduče lokalnog aktora.
	select {
	case pid.Inbox <- envelope:
	default:
		fmt.Printf("[WARN] Send: inbox full for actor %s\n", pid.ID)
	}
}

func (s *ActorSystem) Stop(pid *PID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	close(pid.Inbox)
	delete(s.actors, pid.ID)
	fmt.Println("[INFO] Actor stopped:", pid.ID)
}

func (s *ActorSystem) remoteSend(recipientPID *PID, envelope Envelope) error {
	s.remoteMu.RLock()
	client, ok := s.remoteClients[recipientPID.Address]
	s.remoteMu.RUnlock()

	if !ok {
		conn, err := grpc.Dial(recipientPID.Address, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("failed to dial: %v", err)
		}
		client = &remoteClient{
			client: pb.NewActorServiceClient(conn),
			conn:   conn,
		}
		s.remoteMu.Lock()
		s.remoteClients[recipientPID.Address] = client
		s.remoteMu.Unlock()
	}

	messageBytes, err := serializeMessage(envelope.Msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	remoteEnv := &pb.RemoteEnvelope{
		RecipientId: recipientPID.ID,
		Envelope: &pb.Envelope{
			BinaryMessage: messageBytes,
		},
	}

	// Ako treba primalac da zna ko je pošiljalac
	if envelope.Sender != nil {
		remoteEnv.Envelope.SenderId = envelope.Sender.ID
		remoteEnv.Envelope.SenderAddress = envelope.Sender.Address
	}

	fmt.Println("Hello", remoteEnv)

	_, err = client.client.SendMessage(context.Background(), remoteEnv)
	if err != nil {
		return fmt.Errorf("remote send failed: %v", err)
	}

	return nil
}

func serializeMessage(msg Message) ([]byte, error) {
	return json.Marshal(msg)
}

func deserializeMessage(data []byte) (Message, error) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
