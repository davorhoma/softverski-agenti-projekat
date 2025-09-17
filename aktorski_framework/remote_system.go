package aktorski_framework

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	pb "softverski-agenti-projekat/aktorski_framework/messages"
)

type remoteActorServiceServer struct {
	pb.UnimplementedActorServiceServer
	system *ActorSystem
}

func (s *remoteActorServiceServer) SendMessage(ctx context.Context, in *pb.RemoteEnvelope) (*pb.Empty, error) {
	recipientID := in.GetRecipientId()
	envelopeProto := in.GetEnvelope()

	// Pronađi lokalnog primaoca
	pid, ok := s.system.GetPID(recipientID)
	if !ok {
		return nil, fmt.Errorf("recipient actor %s not found on this node", recipientID)
	}

	msg, err := deserializeMessage(envelopeProto.GetBinaryMessage())
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", err)
	}

	senderPID := &PID{
		ID:      envelopeProto.GetSenderId(),
		Address: envelopeProto.GetSenderAddress(),
	}
	envelope := Envelope{
		Msg:    msg,
		Sender: senderPID,
	}

	// Prosledi poruku u inbox lokalnog aktora
	select {
	case pid.Inbox <- envelope:
		return &pb.Empty{}, nil
	default:
		return nil, fmt.Errorf("inbox full for actor %s", recipientID)
	}
}

func StartRemoteServer(system *ActorSystem, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterActorServiceServer(s, &remoteActorServiceServer{system: system})

	fmt.Printf("gRPC server listening at %s\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}
