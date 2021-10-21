package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Block struct {
    Index int
    Timestamp string
    BPM int
    Hash string
    PrevHash string
}

var BlockChain []Block

func calculateHash(block Block) string {
    record:=string(block.Index)+block.Timestamp+string(block.BPM)+block.PrevHash
    h:=sha256.New()
    h.Write([]byte(record))
    hashed:=h.Sum(nil)
    return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block,BPM int) (Block,error) {
    var newBlock Block
    t:=time.Now()
    newBlock.Index=oldBlock.Index+1
    newBlock.Timestamp=t.String()
    newBlock.PrevHash=oldBlock.Hash
    newBlock.Hash=calculateHash(newBlock)
    return newBlock,nil
}

func isBlockValid(newBlock Block,oldBlock Block) bool {
    if oldBlock.Index+1!=newBlock.Index {
        return false
    }
    if oldBlock.Hash!=newBlock.PrevHash{
        return false
    }
    if calculateHash(newBlock)!=newBlock.Hash {
        return false
    }
    return true
}

func replaceChain(newBlocks []Block) {
    if len(newBlocks)>len(BlockChain) {
        BlockChain=newBlocks
    }
}

//Http Server
type Message struct {
    BPM int
}

func makeMuxRouter() http.Handler {
    muxRouter:=mux.NewRouter()
    muxRouter.HandleFunc("/",hanldeGetBlockchain).Methods("GET")
    muxRouter.HandleFunc("/",handleWriteBlock).Methods("POST")
    return muxRouter
}

func respondWithJson(w http.ResponseWriter,r *http.Request,code int,payload interface{}) {
    response,err:=json.MarshalIndent(payload,"" ," ")
    if err!=nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("HTTP 500: Internal Server Error"))
        return
    }
    w.WriteHeader(code)
    w.Write(response)
}

func hanldeGetBlockchain(w http.ResponseWriter,r *http.Request) {
    bytes,err:=json.MarshalIndent(BlockChain,""," ")
    if err!=nil {
        http.Error(w,err.Error(),http.StatusInternalServerError)
        return
        
    }
    io.WriteString(w,string(bytes))
}

func handleWriteBlock(w http.ResponseWriter,r *http.Request) {
    var m Message
    decoder:=json.NewDecoder(r.Body)
    if err:=decoder.Decode(&m);err!=nil {
        respondWithJson(w,r,http.StatusBadRequest,r.Body)
    }
    defer r.Body.Close()
    newBlock,err:=generateBlock(BlockChain[len(BlockChain)-1],m.BPM)
    if err!=nil {
        respondWithJson(w,r,http.StatusInternalServerError,m)
        return
    }
    if isBlockValid(newBlock,BlockChain[len(BlockChain)-1]) {
        newBlockChain:=append(BlockChain,newBlock)
        replaceChain(newBlockChain)
        spew.Dump(BlockChain)
    }
    respondWithJson(w,r,http.StatusCreated,newBlock)
}
func run() error {
    mux:=makeMuxRouter()
    httpAddr:=os.Getenv("PORT")
    log.Println("Listening on",os.Getenv("PORT"))
    s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err:=s.ListenAndServe(); err!=nil {
	    return err
	}
	return nil
}

func main() {
    err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(genesisBlock)
		BlockChain = append(BlockChain, genesisBlock)
	}()
	log.Fatal(run())
}