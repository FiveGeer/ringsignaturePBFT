package network

import (
	"github.com/ringPBFT/pbft/mylog"
	pb "github.com/ringPBFT/pbft/proto_pbft"
	"strconv"
	"fmt"
)

func (node *Node) StartViewChange() {
	mylog.Info.Println("################Send ViewChange message!!##################")

	viewChangeMsg := &pb.ViewChangeMsg{
		ViewID: node.View.ID,
		NodeID: node.NodeID,
		PrimaryNode: strconv.Itoa(node.View.Primary),
	}
	Broadcast(viewChangeMsg, node.Port_server, node.AllNodetable, FailNodeTable)
}

func (node *Node) StartNewView(){
	node.View.ID += 1
	node.View.Primary = (node.View.Primary + 1) % node.ConsensusData.TotalNum
	mylog.Info.Printf("new ViewID = %d, new Primary = %d\n", node.View.ID, node.View.Primary)
}

func (node *Node) HeartbeatDeal() {
	for {
		select {
		case port := <-FailPort:
			fmt.Println(port)
			fmt.Println(node.View.Primary)
			if port == node.View.Primary + 50000{
				addr := "localhost:" + strconv.Itoa(port)
				mylog.Info.Printf("HeartbeatTest Primary node, address is %s\n", addr)
				PrimaryNodeFail := HeartbeatT(addr, node.ConsensusData.ViewChangeTimeOut)
				if !PrimaryNodeFail && !FailNodeTable[addr]{
					fmt.Printf("TCP %s error\n", addr)
					FailNodeTable[addr] = true
					node.StartViewChange()
				}
			}else{
				addr := "localhost:" + strconv.Itoa(port)
				mylog.Info.Printf("HeartbeatTest normal node, address is %s\n", addr)
				NormalNodeFail := HeartbeatT(addr, node.ConsensusData.Timeout)
				if !NormalNodeFail && !FailNodeTable[addr]{
					fmt.Printf("TCP %s error\n", addr)
					FailNodeTable[addr] = true
				}
			}
		}
	}
}