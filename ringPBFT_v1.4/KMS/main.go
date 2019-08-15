package main

import (

	"github.com/ringPBFT/KMS/generate"

	"goji.io"
	"goji.io/pat"
	"fmt"
	"net/http"
	"encoding/json"

	"strconv"
)

var private []generate.Private
var public []generate.Public

func init(){
	private, public = generate.Generate(8)
}

func main(){
		mux := goji.NewMux()
		mux.HandleFunc(pat.Get("/GetPrivateKey"), GetPrivKey)
		mux.HandleFunc(pat.Get("/GetPublicKey"), GetPubKey)
		mux.HandleFunc(pat.Get("/GetPublicKeyRing"), GetPubKeyRing)
		fmt.Println("begining!!")
		http.ListenAndServe("localhost:8888", mux)
}

func GetPrivKey(w http.ResponseWriter, r *http.Request){
	var temp string
	r.ParseForm()
//	num := binary.BigEndian.Uint64(r.Form["number"])
	for _, v := range r.Form["number"]{
		temp += v
	}
	fmt.Println(temp)
	num, err := strconv.Atoi(temp)
	if err != nil{
		fmt.Println(err.Error())
		w.WriteHeader(500)
	}
	pri, _ := json.Marshal(private[num])
	w.Header().Set("Content-Type", "application/json")
	w.Write(pri)
}

func GetPubKeyRing(w http.ResponseWriter, r *http.Request){
	pub, _ := json.Marshal(public)
	w.Header().Set("Content-Type", "application/json")
	w.Write(pub)
}

func GetPubKey(w http.ResponseWriter, r *http.Request){
	var temp string
	r.ParseForm()
//	num := binary.BigEndian.Uint64(r.Form["number"])
	for _, v := range r.Form["number"]{
		temp += v
	}
	fmt.Println(temp)
	num, err := strconv.Atoi(temp)
	if err != nil{
		fmt.Println(err.Error())
		w.WriteHeader(500)
	}
	pub, _ := json.Marshal(public[num])
	w.Header().Set("Content-Type", "application/json")
	w.Write(pub)
}