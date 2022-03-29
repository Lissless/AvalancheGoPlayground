package utxo

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
)

type UTXO struct {
	codecID   string
	txID      string
	UTXOIndex string
	AssetID   string
	Output    string
}

type Output struct {
	typeID       string
	amount       string
	lockTime     string
	threshold    string
	no_addresses string
	addresses    []string
}

type avaUTXO struct {
	codecID   bytes.Buffer
	txID      bytes.Buffer
	UTXOIndex bytes.Buffer
	AssetID   bytes.Buffer
	Output    Output
}

type avaOutput struct {
	typeID       string
	amount       bytes.Buffer
	amountValue  string //big.Int --> https://medium.com/orbs-network/big-integers-in-go-14534d0e490d
	lockTime     bytes.Buffer
	threshold    bytes.Buffer
	no_addresses bytes.Buffer
	addresses    []string
}

// This is for a cb58 string
func decodeUTXO(utxo string) { //} (avax.UTXO){
	utxoBytes, _ := formatting.Decode(formatting.CB58, utxo)
	codecID := utxoBytes[:2]
	transactionID := utxoBytes[2:34]
	outputIndex := utxoBytes[34:38]
	assetID := utxoBytes[38:70]
	utxoOutput := utxoBytes[70:]

	typeID := utxoOutput[:4]
	amount := utxoOutput[4:12]
	locktime := utxoOutput[12:20]
	threshold := utxoOutput[20:24]
	addressAmount := utxoOutput[24:28]
	addresses := utxoOutput[28:]
	//test := utxoOutput[48:]

	da, ee := formatting.FormatAddress("P", constants.GetHRP(1), addresses)
	amountInt := new(big.Int)
	amountInt.SetBytes(amount)
	fmt.Println("decoded address: ", da)
	fmt.Println("decoded address error: ", ee)
	fmt.Println("The amount: ", amountInt)

	fmt.Println(assetID, codecID, transactionID, outputIndex, utxoOutput)
	fmt.Println(typeID, amount, locktime, threshold, addressAmount, addresses)

}
