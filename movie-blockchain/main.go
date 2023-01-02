package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var BlockChain Blockchain

type Blockchain struct {
	blocks []*Block
}

type Block struct {
	PrevHash  string
	Pos       int
	Data      MovieCheckout
	TimeStamp string
	Hash      string
}

type MovieCheckout struct {
	MovieID     string `json:"movie_id"`
	Viewer      string `json:"viewer"`
	CheckoutYOR string `json:"checkout_yor"`
	IsGenesis   bool   `json:"is_genesis"`
}

type Movie struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Director string `json:"director"`
	YOR      string `json:"yor"`
}

func (b *Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)
	data := string(b.Pos) + b.TimeStamp + string(bytes) + b.PrevHash

	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prevBlock *Block, checkoutMovie MovieCheckout) *Block {
	block := &Block{}

	block.PrevHash = prevBlock.Hash
	block.Pos = prevBlock.Pos + 1
	block.Data = checkoutMovie
	block.TimeStamp = time.Now().String()
	block.generateHash()

	return block
}

func (b *Block) validHash(hash string) bool {
	b.generateHash()

	return b.Hash == hash
}

func validBlock(prevBlock, block *Block) bool {
	if prevBlock.Hash != block.PrevHash {
		return false
	}

	if !block.validHash(block.Hash) {
		return false
	}

	if prevBlock.Pos+1 != block.Pos {
		return false
	}

	return true
}

func (bc *Blockchain) AddBlock(data MovieCheckout) {

	prevBlock := bc.blocks[len(bc.blocks)-1]

	block := CreateBlock(prevBlock, data)

	if validBlock(prevBlock, block) {
		bc.blocks = append(bc.blocks, block)
	}
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	io.WriteString(w, string(jbytes))
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutItem MovieCheckout

	if err := json.NewDecoder(r.Body).Decode(&checkoutItem); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write block")
		w.Write([]byte("could not write block"))
		return
	}

	BlockChain.AddBlock(checkoutItem)

	resp, err := json.MarshalIndent(checkoutItem, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not write block"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func newMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie

	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create: %v", err)
		w.Write([]byte("could not create new movie"))
		return
	}

	hash := md5.New()
	io.WriteString(hash, movie.Director+movie.YOR)
	movie.ID = fmt.Sprintf("%x", hash.Sum(nil))

	resp, err := json.MarshalIndent(movie, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marhal payload:%v", err)
		w.Write([]byte("could not save movie data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func GenesisBlock() *Block {
	return CreateBlock(&Block{}, MovieCheckout{IsGenesis: true})
}

func newBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func main() {

	BlockChain = *newBlockchain()

	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newMovie).Methods("POST")

	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()

	log.Println("Listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000", r))
}
