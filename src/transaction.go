package main

import "fmt"

const subsidy = 10 // 보상의 양

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Ixid      []byte
	Vout      int
	ScriptSig string
}
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data} // 현재 Txid의 입력은 비어있는 상태로 Vout의 값은 -1이다.
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	// tx.SetID()
	return &tx
}
