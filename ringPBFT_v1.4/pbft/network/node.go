package network

import (
	"github.com/ringPBFT/pbft/consensus"

	KMs "github.com/ringPBFT/KMS/generate"
	"fmt"
	"time"
	"errors"
	pb "github.com/ringPBFT/pbft/proto_pbft"
	"strconv"
	"math/rand"

	"github.com/ringPBFT/pbft/mylog"
	"github.com/ringPBFT/pbft/localconfig"

)
var NodeGroup map[string]int   //标记节点的投票权限
var ServerTable map[string]int  //节点服务地址
var MsgEntrance  chan interface{}
var ReqFromClient chan *pb.RequestMsg
var FailPort	chan int
var FailNodeTable map[string]bool
var PrivateKey KMs.Private	//节点私钥
var PublicKeyRing []KMs.Public	//共识域内所有节点的公钥
var PublicKey KMs.Public
type Node struct {
	NodeID        string
	View          *View
	CurrentState  *consensus.State
	CommittedMsgs []*pb.RequestMsg // kinda block.
	MsgBuffer     *MsgBuffer
	MsgDelivery   chan interface{}
	Alarm         chan bool
	Port_server    int
	Flag           chan bool
	AllNodetable   []int
	PrimaryTable   []int
	ConsensusData  *consensus.Consensus
	FaultNum	   *ToleranceNum
	FileName		string
}

type MsgBuffer struct {
	ReqMsgs        []*pb.RequestMsg
	PrePrepareMsgs []*pb.PrePrepareMsg
	PrepareMsgs    []*pb.VoteMsg
	CommitMsgs     []*pb.VoteMsg
	ViewChangeMsgs []*pb.ViewChangeMsg
}

type View struct {
	ID      int64
	Primary int
}

type ToleranceNum struct {
	FAll int   //All
	FPri int   //Primary
}

const ResolvingTimeDuration = time.Millisecond * 500 // 0.5 second.

func NewNode(nodeID string , port int, conf *localconfig.TopLevel) *Node {
	const viewID = 10000000000 // temporary.
	MsgEntrance = make(chan interface{}, 1000)
	ReqFromClient = make(chan *pb.RequestMsg, 0)
	FailPort = make(chan int, 5)
	node := &Node{
		// Hard-coded for test.
		NodeID: nodeID,

		View: &View{
			ID: viewID,
			Primary: 0,
		},

		// Consensus-related struct
		CurrentState: nil,
		Port_server : port,
		CommittedMsgs: make([]*pb.RequestMsg, 0),
		MsgBuffer: &MsgBuffer{
			ReqMsgs:        make([]*pb.RequestMsg, 0),
			PrePrepareMsgs: make([]*pb.PrePrepareMsg, 0),
			PrepareMsgs:    make([]*pb.VoteMsg, 0),
			CommitMsgs:     make([]*pb.VoteMsg, 0),
			ViewChangeMsgs: make([]*pb.ViewChangeMsg, 0),
		},

		// Channels
		MsgDelivery: make(chan interface{}),
		Alarm: make(chan bool),
		Flag: make(chan bool),
		AllNodetable: func() (a []int){
			for i := 0; i < conf.Pbft.TotalNum; i++{
				a = append(a, i)
			}
			return
		}(),
		PrimaryTable: func() (a []int){
			for i := 0; i < conf.Pbft.PrimaryNum; i++{
				a = append(a, i)
			}
			return
		}(),
		ConsensusData: &consensus.Consensus{
			TotalNum: conf.Pbft.TotalNum,
			PrimaryNum: conf.Pbft.PrimaryNum,
			Timeout: conf.Pbft.TimeOut,
			ViewChangeTimeOut: conf.Pbft.ViewChangeTimeOut,
			FAllNode: ( conf.Pbft.TotalNum - 1 ) / 3,
			FPrimary: ( conf.Pbft.PrimaryNum - 1 ) / 3,
		},
		FileName: conf.Ledger.Path + "/" + nodeID,
	}
	node.newNodeGroup()
	node.newServerTable()
	node.newFailNodeTable()

	//TCP失败处理
	go node.HeartbeatDeal()

	//将客户端的请求消息转发给主节点
	go node.dispatchReq()

	// Start message dispatcher
	go node.dispatchMsg() //监听entrance消息

	// Start alarm trigger
	go node.alarmToDispatcher()

	// Start message resolver
	go node.resolveMsg()   //监听deliver消息

	return node
}


func (node *Node) Reply(msg *pb.ReplyMsg) error {
	// Print all committed messages.
	fmt.Print("[Committed value]: ")
	for _, value := range node.CommittedMsgs {
//		fmt.Printf("Committed value: %s, %d, %s, %d", value.ClientID, value.Timestamp, value.Operation, value.SequenceID)
		fmt.Printf("{id = %s, operation = %s}, ", value.ClientID, value.Operation)
	}
	fmt.Print("\n")

	return nil
}

// GetReq can be called when the node's CurrentState is nil.
// Consensus start procedure for the Primary.
func (node *Node) GetReq(reqMsg *pb.RequestMsg) error {
	LogMsg(reqMsg)

	// Create a new state for the new consensus.
	err := node.createStateForNewConsensus()
	if err != nil {
		return err
	}

	// Start the consensus process.
	prePrepareMsg, err := node.CurrentState.StartConsensus(reqMsg)
	if err != nil {
		return err
	}
	mylog.Info.Printf("SequenceID in prePrepareMsg = %d", prePrepareMsg.SequenceID)
	LogStage(fmt.Sprintf("Consensus Process (ViewID:%d)", node.CurrentState.ViewID), false)

	// Send getPrePrepare message
	if prePrepareMsg != nil {
		Broadcast(prePrepareMsg, node.Port_server, node.AllNodetable, FailNodeTable)
		LogStage("Pre-prepare", true)
	}

	return nil
}

// GetPrePrepare can be called when the node's CurrentState is nil.
// Consensus start procedure for normal participants.
func (node *Node) GetPrePrepare(prePrepareMsg *pb.PrePrepareMsg) error {
	LogMsg(prePrepareMsg)

	// Create a new state for the new consensus.
	err := node.createStateForNewConsensus()
	if err != nil {
		return err
	}

	prePareMsg, err := node.CurrentState.PrePrepare(prePrepareMsg, node.NodeID, NodeGroup, PublicKeyRing, PrivateKey)
	if err != nil {
		return err
	}

	if prePareMsg != nil {
		// Attach node ID to the message
		prePareMsg.NodeID = node.NodeID
		node.CurrentState.MsgLogs.PrepareMsgs = append(node.CurrentState.MsgLogs.CommitMsgs, prePareMsg)
		LogStage("Pre-prepare", true)
		Broadcast(prePareMsg, node.Port_server,  node.PrimaryTable, FailNodeTable)
		if NodeGroup[node.NodeID] == 2 {
			LogStage("Prepare", false)

		}else{
			LogStage("Commit", false)
		}
	}

	return nil
}

func (node *Node) GetPrepare(prepareMsg *pb.VoteMsg) error {
	LogMsg(prepareMsg)

	commitMsg, err := node.CurrentState.Prepare(prepareMsg, NodeGroup, PublicKeyRing, node.ConsensusData.FAllNode)
	if err != nil {
		return err
	}

	if commitMsg != nil {
		// Attach node ID to the message
		commitMsg.NodeID = node.NodeID
		node.CurrentState.MsgLogs.CommitMsgs = append(node.CurrentState.MsgLogs.CommitMsgs, commitMsg)
		LogStage("Prepare", true)
		Broadcast(commitMsg, node.Port_server, node.AllNodetable, FailNodeTable)
		LogStage("Commit", false)
	}

	return nil
}

func (node *Node) GetCommit(commitMsg *pb.VoteMsg) error {
	LogMsg(commitMsg)
	if node.CurrentState == nil {
		return errors.New("consensus is finished")
	}else if node.CurrentState.CurrentStage == consensus.Committed{
		return errors.New("ignore the CommitedMsg")
	}
	replyMsg, committedMsg, err := node.CurrentState.Commit(commitMsg, node.ConsensusData.FAllNode)
	if err != nil {
		return err
	}

	if replyMsg != nil {
		if committedMsg == nil {
			return errors.New("committed message is nil, even though the reply message is not nil")
		}

		// Attach node ID to the message
		replyMsg.NodeID = node.NodeID

		// Save the last version of committed messages to node.
		node.CommittedMsgs = append(node.CommittedMsgs, committedMsg)

		LogStage("Commit", true)
		node.Reply(replyMsg)
		SaveToFile(node.FileName, committedMsg)
		LogStage("Reply", true)
		node.Flag <- true

	}

	return nil
}

func (node *Node) GetReply(msg *pb.ReplyMsg) {
	fmt.Printf("Result: %s by %s\n", msg.Result, msg.NodeID)
}

func (node *Node) createStateForNewConsensus() error {
	// Check if there is an ongoing consensus process.
	if node.CurrentState != nil {
		return errors.New("another consensus is ongoing")
	}

	// Get the last sequence ID
	var lastSequenceID int64
	if len(node.CommittedMsgs) == 0 {
		lastSequenceID = -1
	} else {
		lastSequenceID = node.CommittedMsgs[len(node.CommittedMsgs) - 1].SequenceID
	}
	mylog.Info.Printf("lastSequenceID = %d\n", lastSequenceID)
	// Create a new state for this new consensus process in the Primary
	node.CurrentState = consensus.CreateState(node.View.ID, lastSequenceID, node.ConsensusData)

	LogStage("Create the replica status", true)

	return nil
}


func (node *Node) dispatchReq() {
	for {
		select {
		case msg := <-ReqFromClient:
			go Client(node.View.Primary + 50000, msg, FailPort)
		}
	}
}

func (node *Node) dispatchMsg() {
	for {
		select {
		case msg := <-MsgEntrance:
			err := node.routeMsg(msg)
			if err != nil {
				fmt.Println(err)
				// TODO: send err to ErrorChannel
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				fmt.Println(err)
				// TODO: send err to ErrorChannel
			}
		}
	}
}

func (node *Node) routeMsg(msg interface{}) []error {
	switch msg.(type) {
	case *pb.RequestMsg:
		if node.CurrentState == nil {
			// Copy buffered messages first.
			fmt.Println("enter routeMsg")
			fmt.Println(node.CurrentState)

		//	node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
		//	node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)
			// Send messages.
			node.MsgDelivery <- msg.(*pb.RequestMsg)
		} else {
			fmt.Println("enter routeMsg2")
			node.MsgBuffer.ReqMsgs = append(node.MsgBuffer.ReqMsgs, msg.(*pb.RequestMsg))
		}
	case *pb.PrePrepareMsg:
		if node.CurrentState == nil {
			fmt.Println("PrePre routeMsg")
			fmt.Println(node.CurrentState)

		//	node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
		//	node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)
			// Send messages.
			node.MsgDelivery <- msg.(*pb.PrePrepareMsg)
		} else {
			fmt.Println("PrePre routeMsg2")
			fmt.Println(node.CurrentState)
			node.MsgBuffer.PrePrepareMsgs = append(node.MsgBuffer.PrePrepareMsgs, msg.(*pb.PrePrepareMsg))
		}
	case *pb.VoteMsg:
		if msg.(*pb.VoteMsg).MsgType == consensus.PrepareMsg {
			fmt.Println("receive PrepareMsg")
			if node.CurrentState == nil ||  node.CurrentState.CurrentStage != consensus.PrePrepared {
				node.MsgBuffer.PrepareMsgs = append(node.MsgBuffer.PrepareMsgs, msg.(*pb.VoteMsg))
			} else {
				// Copy buffered messages first.
				msgs := make([]*pb.VoteMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)

				// Append a newly arrived message.
				msgs = append(msgs, msg.(*pb.VoteMsg))

				// Empty the buffer.
				node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)

				// Send messages.
				node.MsgDelivery <- msgs
			}
		} else if msg.(*pb.VoteMsg).MsgType == consensus.CommitMsg {
			if node.CurrentState == nil || node.CurrentState.CurrentStage != consensus.Prepared {
				node.MsgBuffer.CommitMsgs = append(node.MsgBuffer.CommitMsgs, msg.(*pb.VoteMsg))
			} else {
				// Copy buffered messages first.
				msgs := make([]*pb.VoteMsg, len(node.MsgBuffer.CommitMsgs))
				copy(msgs, node.MsgBuffer.CommitMsgs)

				// Append a newly arrived message.
				msgs = append(msgs, msg.(*pb.VoteMsg))

				// Empty the buffer.
				node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)

				// Send messages.
				node.MsgDelivery <- msgs
			}
		}

	case *pb.ViewChangeMsg:
		mylog.Info.Println("receive ViewChangeMsg")
		if msg.(*pb.ViewChangeMsg).ViewID == node.View.ID{
			node.MsgBuffer.ViewChangeMsgs = append(node.MsgBuffer.ViewChangeMsgs, msg.(*pb.ViewChangeMsg))
		}
		primaryAddr := "localhost:" + strconv.Itoa(node.View.Primary + 50000)
		if !FailNodeTable[primaryAddr]{
			//向主节点进行tcp测试
			PrimaryNodeFail := HeartbeatT(primaryAddr, node.ConsensusData.ViewChangeTimeOut)
			if !PrimaryNodeFail{
				FailNodeTable[primaryAddr] = true
				node.StartViewChange()
			}
		}
		mylog.Info.Printf("len of node.MsgBuffer.ViewChangeMsgs is %d\n", len(node.MsgBuffer.ViewChangeMsgs))
		mylog.Info.Printf("node.CurrentState.FaultNum.FAll is %d\n", node.ConsensusData.FAllNode)
		if len(node.MsgBuffer.ViewChangeMsgs) > 2 * node.ConsensusData.FAllNode {
			node.StartNewView()
			node.MsgBuffer.ViewChangeMsgs = make([]*pb.ViewChangeMsg, 0)
		}

	}

	return nil
}

func (node *Node) routeMsgWhenAlarmed() []error {
	if node.CurrentState == nil {
		// Check ReqMsgs, send them.
		if len(node.MsgBuffer.ReqMsgs) != 0 {
			msg := node.MsgBuffer.ReqMsgs[0]
			node.MsgBuffer.ReqMsgs = node.MsgBuffer.ReqMsgs[1:len(node.MsgBuffer.ReqMsgs)]
		//	node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
		//	node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)
			node.MsgDelivery <- msg
		}

		// Check PrePrepareMsgs, send them.
		if len(node.MsgBuffer.PrePrepareMsgs) != 0 {
			msg := node.MsgBuffer.PrePrepareMsgs[0]
			node.MsgBuffer.PrePrepareMsgs = node.MsgBuffer.PrePrepareMsgs[1:len(node.MsgBuffer.PrePrepareMsgs)]
		//	node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
		//	node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)
			node.MsgDelivery <- msg
		}
	} else {
		switch node.CurrentState.CurrentStage {
		case consensus.PrePrepared:
			// Check PrepareMsgs, send them.
			if len(node.MsgBuffer.PrepareMsgs) != 0 {
				fmt.Println("MsgBuffer.PrepareMsgs!= 0")
				msgs := make([]*pb.VoteMsg, len(node.MsgBuffer.PrepareMsgs))
				copy(msgs, node.MsgBuffer.PrepareMsgs)
				node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)
				node.MsgDelivery <- msgs
			}else if len(node.MsgBuffer.CommitMsgs) != 0 && NodeGroup[node.NodeID] == 1{
				if len(node.MsgBuffer.CommitMsgs) != 0 {
					fmt.Println("MsgBuffer.CommitMsgs!= 0")
					msgs := make([]*pb.VoteMsg, len(node.MsgBuffer.CommitMsgs))
					copy(msgs, node.MsgBuffer.CommitMsgs)
					node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
					node.MsgDelivery <- msgs
				}
			}

		case consensus.Prepared:
			// Check CommitMsgs, send them.
			if len(node.MsgBuffer.CommitMsgs) != 0 {
				fmt.Println("MsgBuffer.CommitMsgs!= 0")
				msgs := make([]*pb.VoteMsg, len(node.MsgBuffer.CommitMsgs))
				copy(msgs, node.MsgBuffer.CommitMsgs)
				node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
				node.MsgDelivery <- msgs
			}
		}
	}

	return nil
}

func (node *Node) resolveMsg() {
	for {
		// Get buffered messages from the dispatcher.
		msgs := <-node.MsgDelivery
		switch msgs.(type) {
		case *pb.RequestMsg:
			 err := node.resolveRequestMsg(msgs.(*pb.RequestMsg))
			if err != nil{
					fmt.Println(err)
				}
				// TODO: send err to ErrorChannel
		case *pb.PrePrepareMsg:
			err := node.resolvePrePrepareMsg(msgs.(*pb.PrePrepareMsg))
			if err != nil {
					fmt.Println(err)
				// TODO: send err to ErrorChannel
			}
		case []*pb.VoteMsg:
			voteMsgs := msgs.([]*pb.VoteMsg)
			if len(voteMsgs) == 0 {
				break
			}

			if voteMsgs[0].MsgType == consensus.PrepareMsg {
				errs := node.resolvePrepareMsg(voteMsgs)
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err)
					}
					// TODO: send err to ErrorChannel
				}
			} else if voteMsgs[0].MsgType == consensus.CommitMsg {
				errs := node.resolveCommitMsg(voteMsgs)
				if len(errs) != 0 {
					for _, err := range errs {
						fmt.Println(err)
					}
					// TODO: send err to ErrorChannel
				}

			}
		}
	}
}

func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(ResolvingTimeDuration)
		node.Alarm <- true
	}
}

func (node *Node) resolveRequestMsg(msgs *pb.RequestMsg) error {

	// Resolve messages
		err := node.GetReq(msgs)
		if err != nil {
			return err
		}

	return nil
}

func (node *Node) resolvePrePrepareMsg(msgs *pb.PrePrepareMsg) error {
	// Resolve messages
		err := node.GetPrePrepare(msgs)
		if err != nil {
			return err
		}

	return nil
}

func (node *Node) resolvePrepareMsg(msgs []*pb.VoteMsg) []error {
	errs := make([]error, 0)

	// Resolve messages
	for _, prepareMsg := range msgs {
		err := node.GetPrepare(prepareMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (node *Node) resolveCommitMsg(msgs []*pb.VoteMsg) []error {
	errs := make([]error, 0)
	// Resolve messages
	for _, commitMsg := range msgs {
		err := node.GetCommit(commitMsg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func Broadcast(msg interface{}, portServer int, table []int, failNodeTable map[string]bool) {
	for _, v := range table {
		if ServerTable[strconv.Itoa(v)] != portServer {
			addr := "localhost:" + strconv.Itoa(ServerTable[strconv.Itoa(v)])
			if !failNodeTable[addr] {
				fmt.Printf("Broadcast port = %d\n", ServerTable[strconv.Itoa(v)])
				Client(ServerTable[strconv.Itoa(v)], msg, FailPort)
			}
		}
	}
}

func (node *Node) ClearState()  {
	for {
		select {
		case <-node.Flag:
			node.MsgBuffer.CommitMsgs = make([]*pb.VoteMsg, 0)
			node.MsgBuffer.PrepareMsgs = make([]*pb.VoteMsg, 0)
			node.CurrentState = nil

		}
	}

}

func (node *Node) newServerTable() {
	ServerTable = make(map[string]int, node.ConsensusData.TotalNum)
	for i := 0; i < node.ConsensusData.TotalNum; i++{
		ServerTable[strconv.Itoa(i)] = 50000 + i
	}
}

func (node *Node) newNodeGroup(){
	rand.Seed(time.Now().UnixNano())
	NodeGroup = make(map[string]int, node.ConsensusData.TotalNum)
	for i := 0; i < node.ConsensusData.TotalNum; i++{
		switch {
		case i < 4:
			NodeGroup[strconv.Itoa(i)] = 2
/*		case i < 8:
			ServerTable[strconv.Itoa(i)] = 3

		case i < 12:
			ServerTable[strconv.Itoa(i)] = 2*/
		default:
			NodeGroup[strconv.Itoa(i)] = 1
		}
	}
}



func (node *Node) newFailNodeTable(){
	FailNodeTable = make(map[string]bool, node.ConsensusData.TotalNum - 1)
	host := "localhost:"
	for _, v := range node.AllNodetable{
		addr := host + strconv.Itoa( v + 50000 )
		FailNodeTable[addr] = false
	}
}