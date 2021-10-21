package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
    "fmt"
	"github.com/davecgh/go-spew/spew"
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
    newBlock.BPM=BPM
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
func handleConn(conn net.Conn) {
    defer conn.Close()
    io.WriteString(conn,"Enter BPM: ")
    scanner:=bufio.NewScanner(conn)
    go func() {
        for scanner.Scan() {
            bpm,err:=strconv.Atoi(scanner.Text())
            fmt.Println(bpm)
            if err!=nil {
                log.Printf("%v: Not a number %v",scanner.Text(),err)
                continue
            }
            
            newBlock,err:=generateBlock(BlockChain[len(BlockChain)-1],bpm)
            if err!=nil {
                log.Fatal(err)
            }
            if isBlockValid(newBlock,BlockChain[len(BlockChain)-1]) {
                newBlockChain:=append(BlockChain,newBlock)
                replaceChain(newBlockChain)
            }
            bcServer <- BlockChain
            io.WriteString(conn,"\nEnter new BPM: ")
        }
    }()
    go func() {
		for {
			time.Sleep(30 * time.Second)
			output, err := json.Marshal(BlockChain)
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(conn, string(output))
		}
	}()
	for _ = range bcServer {
		spew.Dump(BlockChain)
	}
}
var bcServer chan []Block
func main() {
    err:=godotenv.Load()
    if err!=nil {
        log.Fatal(err)
    }
    bcServer=make(chan []Block)
    t:=time.Now()
    genesisBlock:=Block{0,t.String(),0,"",""}
    spew.Dump(genesisBlock)
    BlockChain=append(BlockChain,genesisBlock)
    server,err:=net.Listen("tcp",":"+os.Getenv("PORT"))
    if err!=nil {
        log.Fatal(err)
    }
    defer server.Close()
    for {
        conn,err:=server.Accept()
        if err!=nil {
            log.Fatal(err)
        }
        go handleConn(conn)
    }
}
