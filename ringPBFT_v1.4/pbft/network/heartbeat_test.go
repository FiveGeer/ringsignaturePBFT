package network

import (
	"testing"

)

func TestHeartbeatClient(t *testing.T) {
	err := HeartbeatClient("localhost:50000")
	if err == nil{
		t.Log("WELL")
	}

}

func TestHeartbeatT(t *testing.T) {
	addr := "localhost:50003"
	flag := HeartbeatT(addr, 5)
	if flag {
		t.Log("true")
	}else {
		t.Log("false")
	}
}
