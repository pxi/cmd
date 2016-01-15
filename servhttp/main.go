// Servhttp serves current working directory through http.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var bind = flag.String("bind", ":8080", "bind to `ADDR`")

func main() {
	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("servhttp: ")

	d, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	srv := http.FileServer(http.Dir(d))
	if err := http.ListenAndServe(*bind, srv); err != nil {
		log.Fatal(err)
	}
}
