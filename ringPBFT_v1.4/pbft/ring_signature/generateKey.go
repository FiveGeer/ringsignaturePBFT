package ring_signature

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strconv"
	KMs "github.com/ringPBFT/KMS/generate"
)

//产生公私钥，编码成json文件存储
var err error
var priv KMs.Private
var pubRing []KMs.Public
var pub KMs.Public
func GetPrivateKey(num int) KMs.Private{
	resp, err := http.Get("http://127.0.0.1:8888/GetPrivateKey?number="+strconv.Itoa(num))
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
 	err = json.NewDecoder(resp.Body).Decode(&priv)
    if err != nil {
    	log.Fatal(err.Error())
    }
	return priv
}

func GetPublicKeyRing() []KMs.Public{
	resp, err := http.Get("http://127.0.0.1:8888/GetPublicKeyRing")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

 	err = json.NewDecoder(resp.Body).Decode(&pubRing)
    if err != nil {
    	log.Fatal(err.Error())
    }
    return pubRing
}

func GetPublicKey(num int) KMs.Public{
	resp, err := http.Get("http://127.0.0.1:8888/GetPublicKey?number="+strconv.Itoa(num))
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

 	err = json.NewDecoder(resp.Body).Decode(&pub)
    if err != nil {
    	log.Fatal(err.Error())
    }
    return pub
}