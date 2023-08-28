package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":9000", "ws address")

func main() {
	flag.Parse()

	streamer := Streamer{}
	bindings := NewServerBindings(streamer)

	http.HandleFunc("/r", bindings.R.ServeHTTP)
	http.HandleFunc("/m", bindings.M.ServeHTTP)
	http.HandleFunc("/p", bindings.P.ServeHTTP)

	fmt.Println("Listening on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
