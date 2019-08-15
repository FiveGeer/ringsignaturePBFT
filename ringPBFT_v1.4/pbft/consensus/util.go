package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	pb "github.com/ringPBFT/pbft/proto_pbft"
	Rs "github.com/ringPBFT/pbft/ring_signature"
	KMs "github.com/ringPBFT/KMS/generate"
	"strconv"
	"log"
)

func Hash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateRingSignature(nodeId string, group int, PublicKeyRing []KMs.Public, NodeGroup map[string]int, PrivateKey KMs.Private) *pb.RingSign {
	var pubRingX []*big.Int
	var pubRingY []*big.Int
	//环签名使用的公钥id集合
	pubId := []string{nodeId}

	var groupMember = []string{nodeId}//保存与nodeID同组的节点ID
	for i, v := range NodeGroup{
		if v == NodeGroup[nodeId] && i != nodeId{
			groupMember = append(groupMember, i)
			pubId = append(pubId, i)
		}
	}
	for _, v := range groupMember{
		id, err := strconv.Atoi(v)
		if err != nil{
			log.Fatal(err.Error())
		}
		pubRingX = append(pubRingX, PublicKeyRing[id].X)
		pubRingY = append(pubRingY, PublicKeyRing[id].Y)
	}

	temp := new(big.Int).SetInt64(int64(group))
	c0, r, yDx, yDy := Rs.RingSignature(PrivateKey.D, pubRingX, pubRingY, temp.Bytes())
	return &pb.RingSign{
		C0: c0,
		R: r,
		YDashX: yDx,
		YDashY: yDy,
		Value: strconv.Itoa(group),
		PubId: pubId,
	}
}

func VerifyRingSignature(sign *pb.RingSign, NodeGroup map[string]int, PublicKeyRing []KMs.Public) bool{
	pubX := make([]*big.Int, len(sign.PubId))
	pubY := make([]*big.Int, len(sign.PubId))
	//m为投票权值
	m, err := strconv.Atoi(sign.Value)
	if err != nil{
		log.Fatalf("ring signature value is not a int, %s\n", err.Error())
	}
	//验证pubid成员是否权值都为m， 若不是，返回false
	for index, id := range sign.PubId{
		if NodeGroup[id] != m{
			return false
		}
		tempID, err := strconv.Atoi(id)
		if err != nil{
			log.Fatalln(err.Error())
		}
		pubX[index] = PublicKeyRing[tempID].X
		pubY[index] = PublicKeyRing[tempID].Y
	}

	value := new(big.Int).SetInt64(int64(m))
	return Rs.VerifySignature(sign.C0, sign.R, sign.YDashX, sign.YDashY, value.Bytes(), pubX, pubY)

}