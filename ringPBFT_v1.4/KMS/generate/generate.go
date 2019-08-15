package generate

import (
	"math/big"
	"crypto/elliptic"
	"fmt"
	"math/rand"

)

type Public struct {
//	elliptic.Curve
	X  *big.Int
	Y  *big.Int
}

type PublicKeyRing struct {
	Ring []Public
}

type Private struct {
	D *big.Int
}

func Generate(num int) ([]Private, []Public){
	rand.Seed(1)
	privKey := make([]Private, num) //存放私钥
	pubKey := make([]Public, num) //存放公钥
	var curve = elliptic.P256()
	var err error
	for i := range privKey{
		privKey[i].D = new(big.Int)
		pubKey[i].Y = new(big.Int)
		pubKey[i].X = new(big.Int)
//		pubKey[i].Curve = curve
		privKey[i].D.SetInt64((int64(rand.Intn(173113)))) //随机产生私钥
		if err != nil {
			fmt.Println(err.Error())
		}
		pubKey[i].X, pubKey[i].Y = curve.ScalarBaseMult(privKey[i].D.Bytes())
	}
	return privKey, pubKey
}