package consensus

import pb "github.com/ringPBFT/pbft/proto_pbft"


type PBFT interface {
	StartConsensus(request *pb.RequestMsg) (*pb.PrePrepareMsg, error)
	PrePrepare(prePrepareMsg *pb.PrePrepareMsg) (*pb.VoteMsg, error)
	Prepare(prepareMsg *pb.VoteMsg) (*pb.VoteMsg, error)
	Commit(commitMsg *pb.VoteMsg) (*pb.ReplyMsg, *pb.RequestMsg, error)
}