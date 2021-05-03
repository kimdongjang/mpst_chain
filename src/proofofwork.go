package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64 // 9223372036854775807 정도의 숫자.
)

const targetBits = 24

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

/*
작업 증명을 위해 SHA256 해시 알고리즘을 사용하는데, 해시 계산을 위해 경계값을 산출
*/
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}
	return pow
}

/*
해시 계산을 위한 데이터 준비 -> 블록의 필드값과 타겟 및 nonce 값을 병합
*/
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)), // nonce는 해시 캐시에서의 카운터 역할.
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}

/*
실제 작업 증명 함수.
*/
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0 // 작업 증명 루프의 카운터

	fmt.Printf(" Mining the block containing \"%s\"\n", pow.block.Transactions)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data) // 해당 블록 data SHA256 해싱
		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 { // 해싱한 정수와 타겟값과의 비교
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}
