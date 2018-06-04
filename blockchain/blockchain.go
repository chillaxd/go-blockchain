package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type blockchain struct {
	Chain  []block `json:"chain"`
	Length int     `json:"length"`
}

type blockchainnodes struct {
	Nodes []string `json:"nodes"`
}

type transaction struct {
	Amount    float64 `json:"amount,float64"`
	Recipient string  `json:"recipient,hex"`
	Sender    string  `json:"sender,hex"`
	Timestamp string
}

type block struct {
	Index             int
	Previousblockhash string
	Proof             int
	Timestamp         string
	Transactions      []transaction
}

// Blockchainidentifier is basic structure of blockchain
type Blockchainidentifier struct {
	Blocks              []block
	Currenttransactions []transaction
	Nodeidentifier      string
	Nodes               map[string]struct{}
	Uniquenodes         []string
}

// newBlock creates new block in the blockchain and adds it to the chain
// :return: <struct> new created block
func (bi *Blockchainidentifier) newBlock(newblockindex int, newproof int, lastblockhash string) (b block) {
	b.Index = newblockindex
	b.Timestamp = time.Now().UTC().Format("2006-01-02 15:04:05")
	b.Transactions = bi.Currenttransactions
	b.Proof = newproof
	b.Previousblockhash = lastblockhash

	bi.Blocks = append(bi.Blocks, b)
	bi.Currenttransactions = []transaction{}

	return
}

// newTransaction creates a new transaction to go into the next mined Block
// :return: <int> the index of the Block that will hold this transaction
func (bi *Blockchainidentifier) newTransaction(t transaction) (upcomingblockindex int) {
	bi.Currenttransactions = append(bi.Currenttransactions, t)

	lastblock := bi.lastBlock()

	if len(lastblock.Timestamp) == 0 {
		upcomingblockindex = 0
	} else {
		upcomingblockindex = lastblock.Index + 1
	}

	return
}

// blockHasher creates a SHA-256 hash of a Block
// :return: <string> Hash string of the block
func blockHasher(b block) (hashvalue string) {
	if len(b.Timestamp) == 0 {
		hashvalue = strings.Repeat("0", 64)
	} else {
		jsonb, err := json.Marshal(&b)
		if err != nil {
			log.Println("ERROR :", err.Error())
			return
		}

		blockhash := sha256.Sum256([]byte(string(jsonb)))
		hashvalue = hex.EncodeToString(blockhash[:])
	}

	return
}

// proofOfWork is simple Proof of Work Algorithm:
// - Find a number p' such that hash(pp') contains leading 4 zeroes, where p is the previous p'
// - p is the previous proof, and p' is the new proof
// :return: <int> the proof which brings 4 leading zeros in hash
func (bi *Blockchainidentifier) proofOfWork(lastproof int) (newproof int) {
	newproof = 0

	for proofValidator(lastproof, newproof) == false {
		newproof++
	}

	return
}

// proofValidator validates the Proof: Does hash contain 4 leading zeroes?
// :return: <bool> true if correct, false if not.
func proofValidator(lastproof int, currentproof int) (validated bool) {
	guess := strconv.Itoa(lastproof) + strconv.Itoa(currentproof)
	guesshash := sha256.Sum256([]byte(guess))
	guesshashhex := hex.EncodeToString(guesshash[:])

	if guesshashhex[:4] == "0000" {
		validated = true
	} else {
		validated = false
	}

	return
}

// isValid checks if the given transaction is valid or not
// :return: <bool> true if valid, false if not
func (bi *Blockchainidentifier) isValid(t transaction) (valid bool) {
	if t.Amount == 0 || len(t.Recipient) == 0 || len(t.Sender) == 0 {
		valid = false
	} else {
		valid = true
	}
	return
}

// lastBlock returns the last block in the chain
// :return: <struct> last block in the chain
func (bi *Blockchainidentifier) lastBlock() (b block) {
	chainlen := len(bi.Blocks)
	if chainlen > 0 {
		b = bi.Blocks[chainlen-1]
	}

	return
}

// getUniqueNodes extracts unique nodes insode blockchain network
func (bi *Blockchainidentifier) getUniqueNodes(bn *blockchainnodes) {
	// Getting all the nodes from api body to blokchain, omiting the duplicates
	for _, node := range bn.Nodes {
		bi.registerNode(node)
	}

	bi.Uniquenodes = []string{}
	// Getting the list of unique node values
	for node := range bi.Nodes {
		bi.Uniquenodes = append(bi.Uniquenodes, node)
	}
}

// registerNode adds a new node to the list of nodes of blockchain
// map is used instead of slice as nodes should contain only unique node
func (bi *Blockchainidentifier) registerNode(address string) {
	u, err := url.Parse(address)
	if err != nil {
		log.Fatal("ERROR :", err.Error())
	}
	bi.Nodes[u.Host] = struct{}{}
}

// blockchainValidator determine if a given blockchain is valid
// :return: <bool> true if valid, false if not
func blockchainValidator(bc []block) bool {
	previousblock := bc[0]
	blocklength := len(bc)
	currentindex := 1
	var currentblock block

	for currentindex < blocklength {
		currentblock = bc[currentindex]

		// Check that the hash of the block is correct
		if currentblock.Previousblockhash != blockHasher(previousblock) {
			return false
		}

		// Check that the Proof of Work is correct
		if proofValidator(previousblock.Proof, currentblock.Proof) == false {
			return false
		}

		previousblock = currentblock
		currentindex++
	}

	return true
}

// blockchainConflictResolver is Consensus Algorithm, it resolves conflicts
// by replacing our chain with the longest one in the network.
// :return: <bool> true if our chain was replaced, false if not
func (bi *Blockchainidentifier) blockchainConflictResolver() bool {
	newblockchain := []block{}
	myblockchainlength := len(bi.Blocks)

	for _, node := range bi.Uniquenodes {
		resp, err := http.Get("http://" + node + "/chain")
		if err != nil {
			log.Println("ERROR :", err.Error())
			return false
		}

		var bc blockchain

		if resp.StatusCode == 200 {
			err := json.NewDecoder(resp.Body).Decode(&bc)
			if err != nil {
				log.Println("ERROR: ", err.Error())
				return false
			}

			l := bc.Length
			c := bc.Chain

			if l > myblockchainlength && blockchainValidator(c) {
				myblockchainlength = l
				newblockchain = c
			}
		}
	}

	if len(newblockchain) != 0 {
		bi.Blocks = newblockchain
		return true
	}

	return false
}
