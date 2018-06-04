package blockchain

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Mine calculates the hash and build the block with transactions
func (bi *Blockchainidentifier) Mine(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr + " GET /mine")

	lastblock := bi.lastBlock()
	lastblockhash := blockHasher(lastblock)
	lastproof := lastblock.Proof
	newproof := bi.proofOfWork(lastproof)
	newblockindex := bi.newTransaction(transaction{
		Amount:    1,
		Recipient: bi.Nodeidentifier,
		Sender:    "0",
		Timestamp: time.Now().UTC().Format("2006-01-02 15:04:05"),
	})

	blockforged := bi.newBlock(newblockindex, newproof, lastblockhash)

	responseMessage := map[string]interface{}{
		"message":             "New Block Forged",
		"index":               blockforged.Index,
		"previous_block_hash": blockforged.Previousblockhash,
		"proof":               blockforged.Proof,
		"transactions":        blockforged.Transactions,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseMessage)
}

// SaveTransaction saves the posted transaction into the chain,
// which is going to be mined
func (bi *Blockchainidentifier) SaveTransaction(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr + " POST /transactions/new")

	var t transaction

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "ERROR: "+err.Error(), 500)
		return
	}

	if bi.isValid(t) == false {
		http.Error(w, "ERROR: Missing values in transaction", 400)
		return
	}

	t.Timestamp = time.Now().UTC().Format("2006-01-02 15:04:05")

	newblockindex := bi.newTransaction(t)

	responseMessage := map[string]string{
		"message": "Transaction will be added in Block#" + strconv.Itoa(newblockindex),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseMessage)
}

// GetChain gives full list of blockchain
func (bi *Blockchainidentifier) GetChain(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr + " GET /chain")

	responseMessage := map[string]interface{}{
		"chain":  bi.Blocks,
		"length": len(bi.Blocks),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseMessage)
}

// RegisterNodes register nodes in the decentralised blockchain network
func (bi *Blockchainidentifier) RegisterNodes(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr + " POST /nodes/register")

	var n blockchainnodes

	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		http.Error(w, "ERROR: "+err.Error(), 500)
		return
	}

	if len(n.Nodes) < 1 {
		http.Error(w, "ERROR: Please supply a valid list of nodes", 400)
		return
	}

	bi.getUniqueNodes(&n)

	responseMessage := map[string]interface{}{
		"message":     "Register nodes in the blockchain",
		"total_nodes": bi.Uniquenodes,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseMessage)
}

// Consensus resolves conflicts in the blockchain network
func (bi *Blockchainidentifier) Consensus(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr + " GET /nodes/resolve")

	bcreplaced := bi.blockchainConflictResolver()
	var message string

	if bcreplaced {
		message = "Our chain has been replaced"
	} else {
		message = "Our chain is authoritative"
	}

	responseMessage := map[string]interface{}{
		"message": message,
		"chain":   bi.Blocks,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseMessage)
}
