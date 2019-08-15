package ring_signature

import (
	"math/big"
	"crypto/elliptic"
	"math/rand"
	crand "crypto/rand"
	"testing"
)

func TestSign(t *testing.T) {
	rand.Seed(1)
	privKey := make([]Private,5) //存放私钥
	pubKey := make([]Public, 5) //存放公钥
	var curve = elliptic.P256()
	var err error
	for i := range privKey{
		privKey[i].D = new(big.Int)
		pubKey[i].Y = new(big.Int)
		pubKey[i].X = new(big.Int)
		pubKey[i].Curve = curve
//		privKey[i].D, err = randFieldElement(curve, rand)
		privKey[i].D.SetInt64((int64(rand.Intn(173113)))) //随机产生私钥
		if err != nil {
			t.Fatal(err.Error())
		}
		pubKey[i].X, pubKey[i].Y = curve.ScalarBaseMult(privKey[i].D.Bytes())
	}
	//公钥环
	pubkeyRing := &PublicKeyRing{
		pubKey,
	}
/*##############################使用环外成员##############################
	var testPriv Private
	testPriv.D = new(big.Int).SetInt64((int64(rand.Intn(173311))))
	spew.Dump(testPriv)
*/
	m := new(big.Int).SetInt64(193127)
	rs, err := sign(crand.Reader, pubkeyRing, m.Bytes(), privKey[1], 2)
	if verify(rs, pubkeyRing, m.Bytes()){
		t.Log("true")
	}else{
		t.Fatal("false")
	}
}
