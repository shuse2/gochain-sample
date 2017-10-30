package main

import (
	"flag"
	"fmt"
	"github.com/shuse2/gochain"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8080", "http port to run server")
	flag.Parse()

	blockchain := gochain.NewBlockchain()
	nodeID := gochain.PseudoUUID()

	log.Printf("Starting gochain http server on %q with %s", *port, nodeID)

	http.Handle("/", gochain.NewHandler(blockchain, nodeID))
	http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
}
