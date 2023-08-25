package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":9000", "ws address")

func main() {
	flag.Parse()

	streamer := Streamer{}
	bindings := NewServerBindings(streamer)

	http.HandleFunc("/m", bindings.M.ServeHTTP)
	http.HandleFunc("/p", bindings.P.ServeHTTP)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
