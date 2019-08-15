package ring_signature

import (
	"testing"
	"math/big"
)

func TestStringToBigInt(t *testing.T){
	var s string = "1234355474523413423354363452345435345"
	b := new(big.Int)
	b, err := b.SetString(s, 0)

	if !err {
			t.Fatal("SetString error")
	}
	t.Log(b)



}
