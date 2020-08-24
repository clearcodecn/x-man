package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	//data, _ := ioutil.ReadFile("./rootCA.crt")
	data, _ := ioutil.ReadFile("./cert.pem")
	block, _ := pem.Decode(data)
	//fmt.Println(block)
	pri, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(pri.KeyUsage)
}
