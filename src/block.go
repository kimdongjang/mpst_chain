package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/March/2021 Chancellor on brink of second bailout for banks"

type Block struct {
	Timestamp    int64 // 블록 생성 시간
	Transactions []*Transaction
	// Data          []byte // 실제 가치의 정보
	PrevBlockHash []byte // 이전 블록의 해시 값
	Hash          []byte // 블록 헤더
	Nonce         int
}

type Blockchain struct {
	tip []byte
	db  *bolt.DB
	// blocks []*Block
}

func (b *Block) SetHash() {
	// timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	// headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{}) // 이전 블록의 해시값 + Data + 블록 생성 시간으로 header를 생성하고 Hash를 진행한다.
	// hash := sha256.Sum256(headers)
	// b.Hash = hash[:]
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	// block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	// block.SetHash()
	// return block
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	// prevBlock := bc.blocks[len(bc.blocks)-1]
	// newBlock := NewBlock(data, prevBlock.Hash)
	// bc.blocks = append(bc.blocks, newBlock)
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error { // BoltDB의 읽기전용 트랜잭션. 새로운 블록의 해시를 채굴하기 위해 DB로부터 마지막 블록의 해시값을 가져옴
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash) // 채굴이 끝난 마지막 해시값으로 블록 생성

	err = bc.db.Update(func(tx *bolt.Tx) error { // 새로운 블록의 해시를 저장하는 l키를 업데이트 한다.
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		if err != nil {
			log.Panic(err)
		}
		return nil
	})

}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
	// return NewBlock("Genesis Block", []byte{})
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

/*
	1. DB 파일 열기
	2. 저장된 블록체인 확인
	3. 블록체인지 존재할 경우
		1) 새로운 블록체인 인스턴스 생성
		2) 블록체인 인스턴스의 끝부분을 DB에 저장된 마지막 블록의 해시로 설정
	4. 블록체인이 존재하지 않을 경우
		1) 제네시스 블록을 생성
		2) DB에 저장
		3) 제네시스 블록의 해시를 마지막 블록의 해시로 저장
		4) 제네시스 블록을 끝부분으로 하는 새로운 블록체인 인스턴스 생성
*/
// func NewBlockchain() *Blockchain {
// 	var tip []byte
// 	db, err := bolt.Open(dbFile, 0600, nil) // boltDB를 여는 표준 방식

// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	err = db.Update(func(tx *bolt.Tx) error { // 읽기-쓰기 트랜잭션가 가능한 db.Update를 열기
// 		b := tx.Bucket([]byte(blocksBucket))
// 		if b == nil {
// 			genesis := NewGenesisBlock()
// 			b, err := tx.CreateBucket([]byte(blocksBucket))
// 			err = b.Put(genesis.Hash, genesis.Serialize())
// 			err = b.Put([]byte("l"), genesis.Hash) // 여기서의 l은 마지막 블록의 해시값을 의미하는 last의 l이다.
// 			tip = genesis.Hash

// 			if err != nil {
// 				log.Panic(err)
// 			}
// 		} else {
// 			tip = b.Get([]byte("l"))
// 		}
// 		return nil
// 	})

// 	bc := Blockchain{tip, db} // 더이상 모든 블록을 저장하지 않고 체인의 끝 블록의 해시만 저장

// 	return &bc

// 	// return &Blockchain{[] *Block{ NewGenesisBlock() }}
// }

func CreateBlockchain(address, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

/*
	트랜잭션의 해시들을 연결하고 연결된 문자열의 해시값을 가져와 사용
*/
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]

	/*
		a[2:]  // same as a[2 : len(a)]
		a[:3]  // same as a[0 : 3]
		a[:]   // same as a[0 : len(a)]
	*/
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

/*
	직렬화
*/
func (b *Block) Serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

/*
	역직렬화
*/
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
