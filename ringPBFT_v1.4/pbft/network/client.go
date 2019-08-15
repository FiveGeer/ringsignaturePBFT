package network

import (

	"google.golang.org/grpc"
	"log"
	pb "github.com/ringPBFT/pbft/proto_pbft"
	"time"
	"context"
	"github.com/ringPBFT/pbft/consensus"
	"strconv"
	"fmt"
	"github.com/ringPBFT/pbft/mylog"
)


var (

	ServerAddr = "localhost:"
)


func Client(port int, msg interface{}, FailPort chan int) {

	ServerAddrClient := ServerAddr + strconv.Itoa(port)
	fmt.Println(ServerAddrClient)
	conn, err := grpc.Dial(ServerAddrClient, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Fatalf("fail to dial %s: %v\n", ServerAddrClient, err)
	}
	err = HeartbeatClient(ServerAddrClient)
	if err != nil{
		FailPort <- port
		mylog.Error.Printf("fail to TCP %s: %v\n", ServerAddrClient, err)
	}
	client := pb.NewBroadCastClient(conn)
	StartTransaction(client, msg)
}
func StartTransaction(client pb.BroadCastClient, msg interface{}) {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	switch msg.(type) {
	case *pb.RequestMsg:
		_, err := client.GetReqResponse(ctx, msg.(*pb.RequestMsg))
		fmt.Println("client send the *pb.RequestMsg to server")
		if err != nil {
			//	log.Fatalf("%v.StartTransation error!!!  %v: ", client, err)
			mylog.Error.Printf("Ignore error!!! %v.StartTransation error!!!  %v: ", client, err)
		}

	case *pb.PrePrepareMsg:
		_, err := client.GetPrePrepareResponse(ctx, msg.(*pb.PrePrepareMsg))
		fmt.Println("client send the *pb.PrePrepareMsg to server")
		if err != nil {
			//	log.Fatalf("%v.StartTransation error!!!  %v: ", client, err)
			mylog.Error.Printf("Ignore error!!! %v.StartTransation error!!!  %v: ", client, err)
		}

	case *pb.VoteMsg:
		voteMsgs := msg.(*pb.VoteMsg)
		if voteMsgs.MsgType == consensus.PrepareMsg {
			_, err := client.GetPrepareResponse(ctx, msg.(*pb.VoteMsg))
			fmt.Println("client send the *pb.PrepareMsg to server")
			if err != nil {
				//	log.Fatalf("%v.StartTransation error!!!  %v: ", client, err)
				mylog.Error.Printf("Ignore error!!! %v.StartTransation error!!!  %v: ", client, err)
			}

		}
		if voteMsgs.MsgType == consensus.CommitMsg {
			_, err := client.GetCommitResponse(ctx, msg.(*pb.VoteMsg))
			fmt.Println("client send the *pb.CommitMsg to server")
			if err != nil {
				//	log.Fatalf("%v.StartTransation error!!!  %v: ", client, err)
				mylog.Error.Printf("Ignore error!!! %v.StartTransation error!!!  %v: ", client, err)
			}
		}

	case *pb.ViewChangeMsg:
		_, err := client.GetViewChangeResponse(ctx, msg.(*pb.ViewChangeMsg))
		fmt.Println("client send the *pb.ViewChangeMsg to server")
		if err != nil {
			mylog.Error.Printf("Ignore error!!! %v.StartTransation error!!!  %v: ", client, err)
		}
	}
}



