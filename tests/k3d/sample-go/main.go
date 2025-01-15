package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("got request")
		w.WriteHeader(200)
		w.Write([]byte("okey dokey"))
	})
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
