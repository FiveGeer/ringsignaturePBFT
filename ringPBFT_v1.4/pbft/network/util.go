package network

import (
	"os"
	"bufio"
	"fmt"
	pb "github.com/ringPBFT/pbft/proto_pbft"
)

func SaveToFile(filename string, msg *pb.RequestMsg) error{
	var f *os.File
	var err error
	if checkFileIsExist(filename) {
		f, err = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
	} else {
		f, err = os.Create(filename)
		fmt.Println("文件不存在, 创建文件")
	}
	check(err)
	defer f.Close()
	bw := bufio.NewWriter(f)
	bw.WriteString(fmt.Sprintf("Committed value: %s, %d, %s, %d", msg.ClientID, msg.Timestamp, msg.Operation, msg.SequenceID) + "\n")
	bw.Flush()
	return err
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}