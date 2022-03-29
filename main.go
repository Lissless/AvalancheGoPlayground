package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	avajson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/utils/rpc"
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

func decodeContainer(height float64) (string, indexer.Container) {

	client := indexer.NewClient("https://indexer-demo.avax.network", "/ext/index/P/block")
	args := indexer.GetContainer{
		Index:    avajson.Uint64(height),
		Encoding: formatting.CB58,
	}

	ctx := context.Background()
	cont, _ := client.GetContainerByIndex(ctx, &args)
	containerID, _ := formatting.EncodeWithChecksum(formatting.CB58, cont.ID[:])

	return containerID, cont
}

func findAllValidBlockIDs(containerData []byte) string {
	length := len(containerData)
	if length < 32 {
		fmt.Println("invalid length for possible BlockID")
		return ""
	}
	var validBlockIDs []string
	for i := 0; i < length-31; i++ {
		check := containerData[i:(i + 32)]
		stringID, err := formatting.EncodeWithChecksum(formatting.CB58, check)
		if err != nil {
			fmt.Println("fatal error")
			break
		}
		time.Sleep((1 * time.Second))
		fmt.Println("Processing possibility number ", (i + 1), " of ", (length - 31))
		resp := getBlock(stringID)
		if resp.Block != nil {
			validBlockIDs = append(validBlockIDs, stringID)
		}
	}

	valid := ""
	for i := 0; i < len(validBlockIDs); i++ {
		valid = valid + validBlockIDs[i] + ", "
	}

	return valid
}

func findBlockID(containerData []byte, wanted string) (int, int, string) {
	length := len(containerData)
	if length < 32 {
		fmt.Println("invalid length for possible BlockID")
		return 0, 0, ""
	}
	var blockIDs []string
	found := -1

	for i := 0; i < length-31; i++ {
		check := containerData[i:(i + 32)]
		stringID, err := formatting.EncodeWithChecksum(formatting.CB58, check)
		if err != nil {
			fmt.Println("fatal error")
			break
		}
		blockIDs = append(blockIDs, stringID)
		if stringID == wanted {
			found = i + 1
		}

	}

	transactionSlice := containerData[:6]
	transactionID := ""
	for i := 0; i < 6; i++ {
		subject := transactionSlice[i]
		numVal := int(subject)
		transactionID = transactionID + strconv.Itoa(numVal)
	}
	transactionID = `"` + transactionID + `"`

	return (length - 31), found, transactionID
}

func getBlock(blockID string) api.GetBlockResponse {
	requester := rpc.NewEndpointRequester("https://api.avax.network", "/ext/P", "platform")
	ctx := context.Background()
	out := new(api.GetBlockResponse)
	ID, err := ids.FromString(blockID)
	if err != nil {
		return *out
	}
	args := api.GetBlockArgs{
		BlockID:  ID,
		Encoding: formatting.JSON,
	}
	requester.SendRequest(ctx, "getBlock", args, out)
	return *out
}

func makeBlockIDAtCSV(blockID string, numEntries int) {
	csvFile, err := os.Create("containerResearchP58.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	initialRow := []string{"Block ID", "Block Height", "Container ID", "Num of Possible IDs", "Index Appeared At", "Possible Transaction Type", "Source or Destination Chain", "Chain Name", "Exported Outputs or Imported Inputs"}
	if err := csvwriter.Write(initialRow); err != nil {
		log.Fatalln("error writing record to file", err)
	}

	var chain string
	var chainName string
	var expinp string
	for i := 0; i < numEntries; i++ {
		chain = "N/A"
		chainName = "N/A"
		expinp = "N/A"
		resp := getBlock(blockID)
		block := resp.Block.(map[string]interface{})
		if checkNotEmpty(block) {
			chain, chainName = getChainType(block)
			expinp = getExportOrImportType(block)
		}

		height := block["height"].(float64)
		containerID, container := decodeContainer(height)
		possibleIDs, appearedAt, transactionID := findBlockID(container.Bytes, blockID)
		row := []string{blockID, fmt.Sprintf("%f", height), containerID, strconv.Itoa(possibleIDs), strconv.Itoa(appearedAt), transactionID, chain, chainName, expinp}

		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}

		blockID = block["parentID"].(string)
	}
}

func checkNotEmpty(block map[string]interface{}) bool {
	for k := range block {
		if k == "tx" {
			return true
		}
	}
	return false
}

func getChainType(block map[string]interface{}) (string, string) {
	tx := block["tx"].(map[string]interface{})
	txUnsigned := tx["unsignedTx"].(map[string]interface{})
	for k := range txUnsigned {
		if k == "sourceChain" {
			return "source", txUnsigned["sourceChain"].(string)
		} else if k == "destinationChain" {
			return "destination", txUnsigned["destinationChain"].(string)
		}
	}
	return "N/A", "N/A"

}

func getExportOrImportType(block map[string]interface{}) string {
	tx := block["tx"].(map[string]interface{})
	txUnsigned := tx["unsignedTx"].(map[string]interface{})
	for k := range txUnsigned {
		if k == "exportedOutputs" {
			return "Exported Outputs"
		} else if k == "importedInputs" {
			return "Imported Inputs"
		}
	}
	return "N/A"
}

func makeFindRelevantIDsCSV(blockID string, numEntries int) {
	csvFile, err := os.Create("containerValidIDsLarge.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	initialRow := []string{"Block Height", "Container ID", "Found Valid Block IDs"}
	if err := csvwriter.Write(initialRow); err != nil {
		log.Fatalln("error writing record to file", err)
	}

	for i := 0; i < numEntries; i++ {
		resp := getBlock(blockID)
		block := resp.Block.(map[string]interface{})
		height := block["height"].(float64)
		containerID, container := decodeContainer(height)
		fmt.Println("Doing entry ", (i + 1), " of ", numEntries)
		valid := findAllValidBlockIDs(container.Bytes)
		row := []string{fmt.Sprintf("%f", height), containerID, valid}
		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}

		blockID = block["parentID"].(string)
	}

}

func main() {
	// dc := "11Atn7rC6LxAVZ3Dyc8xxB8Uz3LMxzos6TjbojRsav3zcJECB919hqakKzUPUTdTiynWUfpMsL2j1nEbcW4WWP1ZkwG48R6DBvUX4s6MVsCm9CdtpU1WbpwSBiPrMKEsZaZkuheaqJdMMtCeyZ8EyFYvJNLMtwyaS43LTb"
	// decodeUTXO(dc)

	// fmt.Println("start sleeping")
	// time.Sleep((5 * time.Second))
	// fmt.Println("Awoken")

	// makeBlockIDAtCSV("26HeCVKx9viUPp5EdkitRMB2tkdzF62dNW1snx9sZGe461m4V2", 100)
	// makeFindRelevantIDsCSV("2HRsYbDTxyxsE9FXgyw6ZtLSzRjNgdSCXxe85kpTwAcVenrv12", 1)
	makeBlockIDAtCSV("GhZVPpopX84JrKiNNP2AXFwyaMBYbcnzG9TdP7Trf2bxVXgC7", 200)

}
