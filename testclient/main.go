package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	bigFile    = "http://cdn2.clearcode.cn/ps%E5%91%A8%E4%B8%80.rar"
	//simpleFile = "http://cdn2.clearcode.cn/index.html"
	simpleFile = "https://www.baidu.com"
	cli        *http.Client
)

func init() {
	pxyURL, _ := url.Parse("http://127.0.0.1:3344")
	cli = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(pxyURL),
		},
	}
}

func main() {
	for {
		time.Sleep(3 * time.Second)
		testSimpleFile()
	}
	//testBigFile()
}

// Transfer-Encoding:chunked
// Content-Length

// 1. Content-Length 存在，则内容必须要和这个长度一致.
// 2. Transfer-Encoding 如果存在这个头，那么Content-Length就不会存在，即使存在 也会被忽略
// 3.

func testBigFile() {
	resp, err := cli.Get(bigFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	for k := range resp.Header {
		fmt.Println(k, ":", resp.Header.Get(k))
	}
	io.Copy(ioutil.Discard, resp.Body)
}

func testSimpleFile() {
	resp, err := cli.Get(simpleFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	for k := range resp.Header {
		fmt.Println(k, ":", resp.Header.Get(k))
	}
	io.Copy(os.Stdout, resp.Body)
}
