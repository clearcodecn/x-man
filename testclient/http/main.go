package main

import (
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/any", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-x509-ca-cert")
		w.Header().Add("Content-Disposition", `attachment; filename="cert.pem"`)
		data, _ := ioutil.ReadFile("./rootCA.crt")
		w.Write(data)
	})
	http.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-x509-ca-cert")
		w.Header().Add("Content-Disposition", `attachment; filename="cert.pem"`)
		data, _ := ioutil.ReadFile("./cert.pem")
		w.Write(data)
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		tpl := `<html><a href="/cert">cert</a></html>`
		writer.Write([]byte(tpl))
	})
	http.ListenAndServe(":6633", nil)
}
