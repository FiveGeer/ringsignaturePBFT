package network

import (
	"net"
	"fmt"
	"time"
)

func HeartbeatClient(addr string) error{
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", addr)
    _, err := net.DialTCP("tcp", nil, tcpAddr)
    return err
}

func HeartbeatServer(add string){
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", add)
    listener, _ := net.ListenTCP("tcp", tcpAddr)
	for {
		conn, _ := listener.Accept()
		go doHandle(conn)
	}
}

func doHandle(conn net.Conn){
	defer conn.Close()
	fmt.Println("server")
}

func HeartbeatT(addr string, duration int64) bool{
	timer := time.NewTimer(time.Duration(duration) * time.Second)
	for{
		select {
		case <-timer.C:
			return false
		default:
			err := HeartbeatClient(addr)
			if err == nil {
				return true
			}
		}
	}

}