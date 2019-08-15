package network

import (
	"net"
	"fmt"
	"context"
	"log"

	pb "github.com/ringPBFT/pbft/proto_pbft"
	"google.golang.org/grpc"

	"strconv"

)


type broadcastServer struct {

}

func (s *broadcastServer) GetReqResponse(ctx context.Context, message *pb.RequestMsg) (*pb.PrePrepareMsg, error) {
	MsgEntrance <- message
	fmt.Println("Get ReqResponse message")
	return &pb.PrePrepareMsg{}, nil
}
func (s *broadcastServer)GetPrePrepareResponse(ctx context.Context, message *pb.PrePrepareMsg) (*pb.VoteMsg, error) {
	MsgEntrance <- message
	fmt.Println("Get PrePrepareResponse message")
	return &pb.VoteMsg{}, nil
}
func (s *broadcastServer) GetPrepareResponse(ctx context.Context, message *pb.VoteMsg) (*pb.VoteMsg, error) {
	MsgEntrance <- message
	fmt.Println("Get PrepareResponse Message")
	return &pb.VoteMsg{}, nil
}
func (s *broadcastServer) GetCommitResponse(ctx context.Context, message *pb.VoteMsg) (*pb.ReplyMsg, error) {
	MsgEntrance <- message
	fmt.Println("Get CommitResponse Message")
	return &pb.ReplyMsg{}, nil
}

func (s *broadcastServer) GetViewChangeResponse(ctx context.Context, message *pb.ViewChangeMsg) (*pb.ViewChangeMsg, error) {
	MsgEntrance <- message
	fmt.Println("Get ViewChangeResponse Message")
	return &pb.ViewChangeMsg{}, nil
}

func Server(PortServer int) {
	Port := ":" + strconv.Itoa(PortServer)
	lis, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	b := newServer()
	fmt.Println("server start")
	pb.RegisterBroadCastServer(grpcServer, b)
	grpcServer.Serve(lis)
}

func newServer() *broadcastServer {
	s := &broadcastServer{
	}
	return s
}
