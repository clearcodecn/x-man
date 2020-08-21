package main

import (
	"bytes"
	"fmt"
)

func main() {
	var buf = bytes.NewBuffer([]byte{1, 2, 3, 4})
	var b []byte = make([]byte, 1)
	n, err := buf.Read(b)
	fmt.Println(b[:n], err)

	n, err = buf.Read(b)
	fmt.Println(b[:n])

	fmt.Println("s-<",buf.String())
}
