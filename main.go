package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/chillaxd/go-blockchain/blockchain"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

var bi blockchain.Blockchainidentifier

func main() {
	log.Println("Starting ...")

	bi.Nodeidentifier = strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
	bi.Nodes = make(map[string]struct{})

	router := mux.NewRouter()

	router.HandleFunc("/mine", bi.Mine).Methods("GET")
	router.HandleFunc("/transactions/new", bi.SaveTransaction).Methods("POST")
	router.HandleFunc("/chain", bi.GetChain).Methods("GET")
	router.HandleFunc("/nodes/register", bi.RegisterNodes).Methods("POST")
	router.HandleFunc("/nodes/resolve", bi.Consensus).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", router))
}
