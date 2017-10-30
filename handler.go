package gochain

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func NewHandler(blockchain *Blockchain, nodeID string) http.Handler {
	h := handler{blockchain, nodeID}
	mux := http.NewServeMux()
	mux.HandleFunc("/chain", buildResponse(h.FullChain))
	mux.HandleFunc("/mine", buildResponse(h.Mine))
	mux.HandleFunc("/transactions/new", buildResponse(h.TransactionNew))
	mux.HandleFunc("/nodes/register", buildResponse(h.RegisterNodes))
	mux.HandleFunc("/nodes/resolve", buildResponse(h.ResolveConflicts))
	return mux
}

type handler struct {
	blockchain *Blockchain
	nodeId     string
}

type response struct {
	value      interface{}
	statusCode int
	err        error
}

func buildResponse(h func(io.Writer, *http.Request) response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := h(w, r)
		msg := resp.value
		if resp.err != nil {
			msg = resp.err.Error()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.statusCode)
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			log.Printf("Cannot encode response: %v", err)
		}
	}
}

func methondNotAllowed(method string) response {
	return response{
		nil,
		http.StatusMethodNotAllowed,
		fmt.Errorf("method %s not allowed", method),
	}
}

func (h *handler) FullChain(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodGet {
		return methondNotAllowed(r.Method)
	}
	resp := map[string]interface{}{"chain": h.blockchain.chain, "length": len(h.blockchain.chain)}
	return response{resp, http.StatusOK, nil}
}

func (h *handler) TransactionNew(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodPost {
		return methondNotAllowed(r.Method)
	}
	tx := Transaction{}
	err := json.NewDecoder(r.Body).Decode(&tx)
	index := h.blockchain.NewTransaction(tx)

	resp := map[string]string{
		"message": fmt.Sprintf("Transactino will be added to block %d", index),
	}
	status := http.StatusCreated
	if err != nil {
		status = http.StatusInternalServerError
		err = fmt.Errorf("Failed to add transaction to blockchain")
	}
	return response{resp, status, err}
}

func (h *handler) Mine(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodPost {
		return methondNotAllowed(r.Method)
	}
	lastBlock := h.blockchain.LastBlock()
	lastProof := lastBlock.Proof
	proof := h.blockchain.ProofOfWork(lastProof)
	newTX := Transaction{
		Sender:    "0",
		Recipient: h.nodeId,
		Amount:    1,
	}
	h.blockchain.NewTransaction(newTX)
	block := h.blockchain.NewBlock(proof, "")
	resp := map[string]interface{}{
		"message":       "New block forged",
		"index":         block.Index,
		"transactions":  block.Transactions,
		"proof":         block.Proof,
		"previous_hash": block.PreviousHash,
	}
	return response{resp, http.StatusOK, nil}
}

func (h *handler) RegisterNodes(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodPost {
		return methondNotAllowed(r.Method)
	}
	body := map[string][]string{}
	err := json.NewDecoder(r.Body).Decode(&body)
	for _, node := range body["nodes"] {
		h.blockchain.RegisterNode(node)
	}
	resp := map[string]interface{}{
		"message": "New nodes have been added",
		"nodes":   h.blockchain.nodes.Keys(),
	}
	status := http.StatusCreated
	if err != nil {
		status = http.StatusInternalServerError
		err = fmt.Errorf("Failed to register ndoes")
	}
	return response{resp, status, err}
}

func (h *handler) ResolveConflicts(w io.Writer, r *http.Request) response {
	if r.Method != http.MethodPost {
		return methondNotAllowed(r.Method)
	}
	msg := "chain is authoritative"
	if h.blockchain.ResolveConflicts() {
		msg = "chain was replaced"
	}
	resp := map[string]interface{}{
		"message": msg,
		"chain":   h.blockchain.chain,
	}
	return response{resp, http.StatusOK, nil}
}
