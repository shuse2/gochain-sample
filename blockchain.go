package gochain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Blockchain struct {
	chain               []Block
	currentTransactions []Transaction
	nodes               StringSet
}

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int64         `json:"proof"`
	PreviousHash string        `json:"previous_hash"`
}

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
}

func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		chain:               make([]Block, 0),
		currentTransactions: make([]Transaction, 0),
		nodes:               NewStringSet(),
	}
	bc.NewBlock(100, "1")
	return bc
}

func (bc *Blockchain) RegisterNode(address string) bool {
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	return bc.nodes.Add(u.Host)
}

func (bc *Blockchain) NewBlock(proof int64, previousHash string) Block {
	prevHash := previousHash
	if previousHash == "" {
		prevBlock := bc.chain[len(bc.chain)-1]
		prevHash = blockHash(prevBlock)
	}

	block := Block{
		Index:        int64(len(bc.chain) + 1),
		Timestamp:    time.Now().UnixNano(),
		Transactions: bc.currentTransactions,
		Proof:        proof,
		PreviousHash: prevHash,
	}
	bc.currentTransactions = nil
	bc.chain = append(bc.chain, block)
	return block
}

func (bc *Blockchain) NewTransaction(transaction Transaction) int64 {
	bc.currentTransactions = append(bc.currentTransactions, transaction)
	return bc.LastBlock().Index + 1
}

func (bc *Blockchain) LastBlock() Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) ProofOfWork(lastProof int64) int64 {
	var proof int64 = 0
	for !bc.ValidProof(lastProof, proof) {
		proof += 1
	}
	return proof
}

func (bc *Blockchain) ValidChain(chain *[]Block) bool {
	lastBlock := (*chain)[0]
	currentIndex := 1
	for currentIndex < len(*chain) {
		block := (*chain)[currentIndex]
		if block.PreviousHash != blockHash(lastBlock) {
			fmt.Println("prev hash wrong")
			return false
		}
		if !bc.ValidProof(lastBlock.Proof, block.Proof) {
			fmt.Println("not valid proof")
			return false
		}
		lastBlock = block
		currentIndex += 1
	}
	return true
}

func (bc *Blockchain) ResolveConflicts() bool {
	neighbors := bc.nodes
	newChain := make([]Block, 0)
	maxLength := len(bc.chain)
	for _, node := range neighbors.Keys() {
		otherChains, err := findExternalChain(node)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if otherChains.Length > maxLength && bc.ValidChain(&otherChains.Chain) {
			maxLength = otherChains.Length
			newChain = otherChains.Chain
		}
	}
	if len(newChain) > 0 {
		bc.chain = newChain
		return true
	}
	return false
}

func (bc *Blockchain) ValidProof(lastProof, proof int64) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
	fmt.Println(guess)
	guessHash := fmt.Sprintf("%x", sha256.Sum256([]byte(guess)))
	fmt.Println(guessHash)
	return guessHash[:4] == "0000"
}

func blockHash(block Block) string {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, block)
	return fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
}

type blockchainInfo struct {
	Length int     `json:"length"`
	Chain  []Block `json:"chain"`
}

func findExternalChain(address string) (blockchainInfo, error) {
	res, err := http.Get(fmt.Sprintf("http://%s/chain", address))
	if err != nil || res.StatusCode != http.StatusOK {
		return blockchainInfo{}, err
	}
	info := blockchainInfo{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return blockchainInfo{}, err
	}
	fmt.Println("info", info)
	return info, nil
}
