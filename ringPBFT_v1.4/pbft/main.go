package main

import (
	"os"
	"github.com/ringPBFT/pbft/network"
	"strconv"
	pb "github.com/ringPBFT/pbft/proto_pbft"
	"fmt"
	"time"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"encoding/json"
	Rs "github.com/ringPBFT/pbft/ring_signature"
	"log"
	"github.com/ringPBFT/pbft/localconfig"
	"github.com/ringPBFT/pbft/mylog"
)

type Transaction struct {
	ID          string   `json:"id"`
	Operation   string   `json:"operation"`
}

func init(){
	network.PublicKeyRing = Rs.GetPublicKeyRing()
}
// 第一个参数：节点编号 第二个参数：服务地址
func main() {
	nodeID := os.Args[1]
	portServer := os.Args[2]
	nodeId, err := strconv.Atoi(nodeID)
	if err != nil{
		log.Fatal(err.Error())
	}
	//获取当前节点公私钥
	network.PrivateKey = Rs.GetPrivateKey(nodeId)
	network.PublicKey = Rs.GetPublicKey(nodeId)
	conf, _ := localconfig.Load()
	//初始化node结构，并开启共识服务
	PortServer, _ := strconv.Atoi(portServer)
	node := network.NewNode(nodeID, PortServer, conf)
	//定期状态清理
	go node.ClearState()
	fmt.Printf("Port_server = %d\n", PortServer)
	go network.Server(PortServer) //开启监听

	//接收客户端交易请求,转发给主节点
	listenAddr := "localhost:" + strconv.Itoa(60000 + nodeId)
	fmt.Printf("ListenAddr: %s\n", listenAddr)
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/transaction"), HandleTransaction)
	http.ListenAndServe(listenAddr, mux)

}

func HandleTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&transaction)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{message: %q}", "Incorrect body")
		return
	}
	msg := pb.RequestMsg{Timestamp:time.Now().UnixNano(), ClientID:transaction.ID, Operation:transaction.Operation, SequenceID:time.Now().Unix()}
	mylog.Info.Printf("Request massage sequenceID = %d\n", msg.SequenceID)
	network.ReqFromClient <- &msg
//	w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("transaction success"))
}



