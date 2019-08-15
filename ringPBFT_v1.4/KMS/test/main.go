package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"math/big"
	"log"

	"strconv"
)

type Private struct {
	D *big.Int
}

type Public struct {
	X *big.Int
	Y *big.Int
}
var pub2 Public
var priv Private
var pub []Public
func main(){
	var err error
	num := 1
	resp, err := http.Get("http://127.0.0.1:8888/GetPrivateKey?number="+strconv.Itoa(num))
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

 	err = json.NewDecoder(resp.Body).Decode(&priv)
    if err != nil {
    	log.Fatal(err.Error())
    }
	fmt.Println(priv.D)

	resp, err = http.Get("http://127.0.0.1:8888/GetPublicKey?number="+strconv.Itoa(num))
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

 	err = json.NewDecoder(resp.Body).Decode(&pub2)
    if err != nil {
    	log.Fatal(err.Error())
    }
	fmt.Println(pub2.X)


	resp, err = http.Get("http://127.0.0.1:8888/GetPublicKeyRing")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

 	err = json.NewDecoder(resp.Body).Decode(&pub)
    if err != nil {
    	log.Fatal(err.Error())
    }

	fmt.Println(pub[1].X)
}
