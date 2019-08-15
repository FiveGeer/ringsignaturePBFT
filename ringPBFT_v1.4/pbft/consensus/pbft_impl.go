package consensus

import (
	"encoding/json"
	"errors"
	"fmt"
	KMs "github.com/ringPBFT/KMS/generate"
	pb "github.com/ringPBFT/pbft/proto_pbft"
	"github.com/ringPBFT/pbft/mylog"
	"strconv"
	"log"
	"github.com/davecgh/go-spew/spew"
)

type State struct {
	ViewID         int64
	MsgLogs        *MsgLogs
	LastSequenceID int64
	CurrentStage   Stage       //PrePrepare Prepare Commit
}

type Consensus struct {
	TotalNum   int
	PrimaryNum int
	Timeout    int64
	ViewChangeTimeOut int64
	FAllNode   int //All domain tolerate
	FPrimary   int //Primary domain tolerate
}

type MsgLogs struct {
	ReqMsg        *pb.RequestMsg
	PrepareMsgs   []*pb.VoteMsg
	CommitMsgs    []*pb.VoteMsg
}

type Stage int
const (
	Idle        Stage = iota // Node is created successfully, but the consensus process is not started yet.
	PrePrepared              // The ReqMsgs is processed successfully. The node is ready to head to the Prepare stage.
	Prepared                 // Same with `prepared` stage explained in the original paper.
	Committed                // Same with `committed-local` stage explained in the original paper.
)

// f: # of Byzantine faulty node
// f = (n­1) / 3
// n = 4, in this case.

// lastSequenceID will be -1 if there is no last sequence ID.
func CreateState(viewID int64, lastSequenceID int64, con *Consensus) *State {
	return &State{
		ViewID: viewID,
		MsgLogs: &MsgLogs{
			ReqMsg:nil,
			PrepareMsgs:make([]*pb.VoteMsg, 0),
			CommitMsgs:make([]*pb.VoteMsg, 0),
		},
		LastSequenceID: lastSequenceID,
		CurrentStage: Idle,
	}
}

func (state *State) StartConsensus(request *pb.RequestMsg) (*pb.PrePrepareMsg, error) {
	// `sequenceID` will be the index of this message.
	mylog.Info.Println("##############startConsensus##############")
	sequenceID := request.SequenceID

	// Find the unique and largest number for the sequence ID
	if state.LastSequenceID != -1 {
		for state.LastSequenceID >= sequenceID {
			sequenceID += 1
		}
	}

	// Assign a new sequence ID to the request message object.
	request.SequenceID = sequenceID
	mylog.Info.Printf("Request massage sequenceID = %d\n", request.SequenceID)
	// Save ReqMsgs to its logs.
	state.MsgLogs.ReqMsg = request

	// Get the digest of the request message
	digest, err := digest(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Change the stage to pre-prepared.
	state.CurrentStage = PrePrepared

	return &pb.PrePrepareMsg{
		ViewID: state.ViewID,
		SequenceID: sequenceID,
		Digest: digest,
		RequestMsg: request,
	}, nil
}

func (state *State) PrePrepare(prePrepareMsg *pb.PrePrepareMsg, nodeId string, NodeGroup map[string]int, PublicKeyRing []KMs.Public, PrivateKey KMs.Private) (*pb.VoteMsg, error) {
	// Get ReqMsgs and save it to its logs like the primary.
	state.MsgLogs.ReqMsg = prePrepareMsg.RequestMsg

	// Verify if v, n(a.k.a. sequenceID), d are correct.
	if !state.verifyMsg(prePrepareMsg.ViewID, prePrepareMsg.SequenceID, prePrepareMsg.Digest) {
		return nil, errors.New("pre-prepare message is corrupted")
	}

	// Change the stage to pre-prepared.
	if NodeGroup[nodeId] == 2{
		state.CurrentStage = PrePrepared
	}else{
		state.CurrentStage = Prepared
	}


	//投票过程，产生环签名
	rs := GenerateRingSignature(nodeId, NodeGroup[nodeId], PublicKeyRing, NodeGroup, PrivateKey)
	fmt.Println("GenerateRingSignature finish")
	return &pb.VoteMsg{
		ViewID: state.ViewID,
		SequenceID: prePrepareMsg.SequenceID,
		Digest: prePrepareMsg.Digest,
		MsgType: PrepareMsg,
		RS: rs,
	}, nil
}


func (state *State) Prepare(prepareMsg *pb.VoteMsg, NodeGroup map[string]int, PublicKeyRing []KMs.Public, f int) (*pb.VoteMsg, error){
	if !state.verifyMsg(prepareMsg.ViewID, prepareMsg.SequenceID, prepareMsg.Digest) {
		return nil, errors.New("prepare message is corrupted")
	}
	mylog.Info.Printf("prepareMsg SequenceID = %d\n", prepareMsg.SequenceID)
	rs := prepareMsg.RS
	spew.Dump(rs)
	// 验证消息的环签名
	if VerifyRingSignature(rs, NodeGroup, PublicKeyRing){
		fmt.Println("!!!!!!!!!!!!!!VerifyRingSignature success!!!!!!!!!!")
		// Append msg to its logs
		times, err := strconv.Atoi(rs.Value)
		fmt.Println(rs.Value)
		if err != nil {
			log.Fatal(err.Error())
		}
		for i := 0; i < times; i++{
		//	state.MsgLogs.PrepareMsgs[prepareMsg.NodeID] = prepareMsg
			state.MsgLogs.PrepareMsgs = append(state.MsgLogs.PrepareMsgs, prepareMsg)
		}
	}

	// Print current voting status
	fmt.Printf("[Prepare-Vote]: %d\n", len(state.MsgLogs.PrepareMsgs))

	if state.prepared(f) {
		// Change the stage to prepared.
		state.CurrentStage = Prepared
		fmt.Println("[Status]:Prepared")

		return &pb.VoteMsg{
			ViewID: state.ViewID,
			SequenceID: prepareMsg.SequenceID,
			Digest: prepareMsg.Digest,
			MsgType: CommitMsg,
		}, nil
	}

	return nil, nil
}

func (state *State) Commit(commitMsg *pb.VoteMsg, F int) (*pb.ReplyMsg, *pb.RequestMsg, error) {
	mylog.Info.Printf("commitMsg SequenceID = %d\n", commitMsg.SequenceID)
	if !state.verifyMsg(commitMsg.ViewID, commitMsg.SequenceID, commitMsg.Digest) {
		return nil, nil, errors.New("commit message is corrupted")
	}

	// Append msg to its logs
	//state.MsgLogs.CommitMsgs[commitMsg.NodeID] = commitMsg
	state.MsgLogs.CommitMsgs = append(state.MsgLogs.CommitMsgs, commitMsg)
	// Print current voting status
	fmt.Printf("[Commit-Vote]: %d\n", len(state.MsgLogs.CommitMsgs))

	if state.committed(F) {
		// This node executes the requested operation locally and gets the result.
		result := "Executed"

		// Change the stage to prepared.
		state.CurrentStage = Committed

		return &pb.ReplyMsg{
			ViewID: state.ViewID,
			Timestamp: state.MsgLogs.ReqMsg.Timestamp,
			ClientID: state.MsgLogs.ReqMsg.ClientID,
			Result: result,
		}, state.MsgLogs.ReqMsg, nil
	}

	return nil, nil, nil
}

func (state *State) verifyMsg(viewID int64, sequenceID int64, digestGot string) bool {
	// Wrong view. That is, wrong configurations of peers to start the consensus.
	fmt.Printf("state.ViewID = %d, message viewID = %d\n", state.ViewID, viewID)
	if state.ViewID != viewID {
		mylog.Error.Println("view ID verify error!!!")
		return false
	}

	// Check if the Primary sent fault sequence number. => Faulty primary.
	// TODO: adopt upper/lower bound check.
	if state.LastSequenceID != -1 {
		if state.LastSequenceID >= sequenceID {
			mylog.Error.Printf("LastSequenceID is %d\n", state.LastSequenceID)
			mylog.Error.Printf("SequenceID is %d\n", sequenceID)
			mylog.Error.Println("LastSequenceID verify error!!!")
			return false
		}
	}

	digest, err := digest(state.MsgLogs.ReqMsg)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Check digest.
	if digestGot != digest {
		mylog.Error.Println("digest verify error!!!")
		return false
	}

	return true
}

func (state *State) prepared(f int) bool {
	if state.MsgLogs.ReqMsg == nil {
		return false
	}

	if len(state.MsgLogs.PrepareMsgs) <= 2 * f {
		return false
	}

	return true
}

func (state *State) committed(F int) bool {
/*	if !state.prepared() {
		return false
	}*/

	if len(state.MsgLogs.CommitMsgs) <= 2 * F {
		return false
	}

	return true
}

func digest(object interface{}) (string, error) {
	msg, err := json.Marshal(object)

	if err != nil {
		return "", err
	}

	return Hash(msg), nil
}