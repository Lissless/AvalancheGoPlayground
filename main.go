package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

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

// type avaUTXO struct {
// 	codecID   bytes.Buffer
// 	txID      bytes.Buffer
// 	UTXOIndex bytes.Buffer
// 	AssetID   bytes.Buffer
// 	Output    Output
// }

// type avaOutput struct {
// 	typeID       string
// 	amount       bytes.Buffer
// 	amountValue  string //big.Int --> https://medium.com/orbs-network/big-integers-in-go-14534d0e490d
// 	lockTime     bytes.Buffer
// 	threshold    bytes.Buffer
// 	no_addresses bytes.Buffer
// 	addresses    []string
// }

func decodeUTXO(utxo string) {
	utxoStruct := UTXO{
		codecID:   utxo[:4],
		txID:      utxo[4:68],
		UTXOIndex: utxo[68:76],
		AssetID:   utxo[76:140],
		Output:    utxo[140:],
	}

	codecID := utxo[:4]
	fmt.Println("CodecID: ", codecID)
	utxo = utxo[4:]
	txID := utxo[:64]
	fmt.Println("TransactionID: ", txID)
	utxo = utxo[64:]
	UTXOIndex := utxo[:8]
	fmt.Println("UTXOIndex: ", UTXOIndex)
	utxo = utxo[8:]
	AssetID := utxo[:64]
	fmt.Println("AssetID: ", AssetID)
	utxo = utxo[64:]
	Output := utxo
	fmt.Println("Output: ", Output)
	fmt.Println("UTXO struct: ", utxoStruct)

	dec, _ := strconv.ParseUint(txID, 16, 64)
	fmt.Println("Txid?: ", dec)

	//data := []byte(txID)
	//fmt.Println(string(data))

	decode, oof := hex.DecodeString(txID)
	if oof != nil {
		panic(oof)
	}
	fmt.Println("TransactionID Decoded: ", string(decode))
}

func decodeContainer(height float64, searchID string) (string, int, int, string) {

	client := indexer.NewClient("https://indexer-demo.avax.network", "/ext/index/P/block")
	args := indexer.GetContainer{
		Index:    avajson.Uint64(height),
		Encoding: formatting.CB58,
	}

	ctx := context.Background()
	cont, _ := client.GetContainerByIndex(ctx, &args)
	containerID, _ := formatting.EncodeWithChecksum(formatting.CB58, cont.ID[:])
	// fmt.Println(cont)

	possibleIDs, appearedAt, transactionID := findBlockID(cont.Bytes, searchID)

	// fmt.Println("Container Decoded: ", string(containerBytes))
	return containerID, possibleIDs, appearedAt, transactionID
}

func makeResearchCSV(blockID string, numEntries int) {
	csvFile, err := os.Create("containerResearch1.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	initialRow := []string{"Block ID", "Block Height", "Container ID", "Num of Possible IDs", "Index Appeared At", "Possible Transaction Type"}
	if err := csvwriter.Write(initialRow); err != nil {
		log.Fatalln("error writing record to file", err)
	}
	for i := 0; i < numEntries; i++ {
		resp := getBlock(blockID)
		block := resp.Block.(map[string]interface{})
		height := block["height"].(float64)
		containerID, possibleIDs, appearedAt, transactionID := decodeContainer(height, blockID)
		row := []string{blockID, fmt.Sprintf("%f", height), containerID, strconv.Itoa(possibleIDs), strconv.Itoa(appearedAt), transactionID}

		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}
		blockID = block["parentID"].(string)
		// fmt.Println(containerID, possibleIDs, appearedAt, transactionID)
		//nextID := resp.Block["parentID"]
	}

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

func findBlockID(containerData []byte, wanted string) (int, int, string) {
	length := len(containerData)
	if length < 32 {
		fmt.Println("invalid length for possible BlockID")
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
		// fmt.Println("Possible ID", (i + 1), ": ", stringID)
		blockIDs = append(blockIDs, stringID)
		if stringID == wanted {
			found = i + 1
		}

	}

	// fmt.Println("ID located at Position: ", found)
	// fmt.Println("done")

	transactionSlice := containerData[:6]
	transactionID := ""
	// TODO: This outputs only zeros even if the final byte is a 1, look into how to fix
	for i := 0; i < 6; i++ {
		numVal, _ := strconv.Atoi(string(transactionSlice[i]))
		transactionID = transactionID + strconv.Itoa(numVal)
	}
	transactionID = `"` + transactionID + `"`

	return (length - 31), found, transactionID
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

func dcUTXO(utxo string) { //} (avax.UTXO){
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

	dc, ee := formatting.FormatAddress("P", constants.GetHRP(1), addresses)
	amountInt := new(big.Int)
	amountInt.SetBytes(amount)
	fmt.Println("decoded address: ", dc)
	fmt.Println("decoded address error: ", ee)
	fmt.Println("The amount: ", amountInt)

	fmt.Println(assetID, codecID, transactionID, outputIndex, utxoOutput)
	fmt.Println(typeID, amount, locktime, threshold, addressAmount, addresses)

}

func main() {
	// //s := "000000000001e94973022dc63ddd4cf9f92732dfd7eabc88a4e93fe3d474e691b38d57e6dc8a0000002e000000000002d2bdea152eec9865cd6d60e1f778c2b8d13849190789dfafa0a91249d69d7bc600000000001435d3b538f228"
	// //s := "000000000000a819109ce7e48b4c3631d8b2546d1da34f669e2f59e6887a4cc74b7b6349bdab00000000623253890000000000143578000004a13082049d30820285a003020102020100300d06092a864886f70d01010b050030003020170d3939313233313030303030305a180f32313230303932323133353232345a300030820222300d06092a864886f70d01010105000382020f003082020a0282020100b9a2d7fdc12e76a683b0293bd509962ddaef2885743e68808a70455758f21cc15186ce80b673146c0ce7f8027cc55280f708afcb91a2275e1907edbc0ab49ea06f5034776a95841f303017690ca429b909950193ca3bea1926af19645a5668814933e643e8d86afb8b0a6fb95ee1b8f759f080199804150f2abbabc0268b8c5193edc6d638ab237e0d694dfea54ed24db2cc08f6f13b0b9aba5d43234427ba2a908f8c4094f16c7f69bad24e8b98b6a4acdb286542cc0d7871c693517fe2ef19f9fc920c813fd0adfae84de52cb6118d5f1c138714a769ad6ba0f8e9deaecfc50a0eafeadc3aa0ab9387e5092747d047b83a80c931796663a3cb574532434af1c02b664bbc71a1a9b9126f55a1fac78de03b6c8e7d51866a6305e019f924e85e85a7a1f8bb3cb4ec082edb6ef7256b4ba47bdbe20348ce930b4fc4bac1ee48d1c8e8a0ce894717ba27f8e4069b97600b2d0350f6331aa4461738fc1b95598b0c087221672b0d420c88465c86b254d4403ab4311e35de77ccd1cf5318beeb6f20ae45a1986b571312cb97040df64ce0bba87ec60aca667543a93f793bdbe16ada71286995a3bb7525625848f5c5e7b7ce4341f973bc305a4c33b6d00fdff17c3b823108003fa64316dce8360cdd98cb63e043f44681a19033974875269391b880e88181061b907bae620055ac3dc36cc0ccf0be83e347f8400dfe1c8445168fa30203010001a320301e300e0603551d0f0101ff0404030204b0300c0603551d130101ff04023000300d06092a864886f70d01010b0500038202010081e921ad1d3e56bf7f0dcd649f02efa52e3718ee3eb0e29ddd650e520d4d3853c971c3d88b88e91f883352700ef9e672591aac058e2f92c9046612f05ced2a5068328fc2dd45a856b12646791d3190be2a7ddaff021bef3e108ec8af242d6442ef88f3ef040754dd68d61f7b245369b838bdcccc24cea3a87aab728930a7f12c0c57b254a559c964c5bf7896224e7d45a210ea24e26488926bcf3504703971f22c245488703e967fec3a54271e2005b88edaeb91bc099f2d5c86a4a7a85bc395eca67bf39eddf394bf593ef323db491b4b171bbd767f34820e0b768a16f99102566cbf32b99286859635f0edeff5d0f5866b294eb153d390986f375cb8787986dad49b8d40eae9ab1c10c3989bb5a421218a23145e17e673b7414d6d3a93a8e59f3ea6fc171247e6a54c7929d1260e89c4aff965f6027b8b185cfe2371d6c62255729455493f52f7e9511d91132f7b1d8a7897cc237bfabc62492896153fcd230ffa01c58a1b6e3c2280c75333162dbfdf2da9f10bd298548dd38a237111d621536e5d638d312b1f39918beb79260ff250212a1baaf72d3fa0925e57543371a2663c37f53f58b96b0857540c0b2140dd212d3ed15ce4799731d568e58b63e68cf1782e25c4ae423ca5123d2802339c3239fcb16fc5c541649d6f59a2f22f2b9d345c19c2f5e1d12de40f8e7e7235de73e6ec796ef57721c6359c9c7f9306cf690000017f000000000003a7c905c528007f9667dde29ee7ea279e20f44eae9cee73845f7ce190b366aab4000000000014367900000001000000110000000100000000000000000000000000000000000000000000000000000000000000000000000121e67317cbc4be2aeb00677ad6462778a8f52274b9d605df2591b23027a87dff000000070000008763bf1f7c000000000000000000000001000000014ef75cd9f419fa19c1837eed1eefdc23bed105500000000000000000ed5f38341e436e5d46e2bb00b45d62ae97d1b050c64bc634ae10626739e35c4b00000001838d3590d39b1e44c4bf9d78de81c19a948a0a6d96dda49b66a806fd46cc0caa0000000021e67317cbc4be2aeb00677ad6462778a8f52274b9d605df2591b23027a87dff000000050000008763ce61bc0000000100000000000000010000000900000001cfa9e30f0f274d97d41d944fed76924f1b9e3eb8e16d052b39eb5db99d80969f7ecdcda6d04ca3fab20d782218e65b28a3559393f84729ec7f1ef50657f3735d0000000200025b663f51be24c5aa989dfc10755f2e85725fec297c8b9546fcd05ef7c37b7b48c85eed306e2ce855d4fb9521cbb5eaee8cfa20ae7245e13bc1e94a1320188cbcac03c4002390c62a8423be1daf127397e6f989e755792b5ecc574436b909834fed52c4001ca8555c705dac259c299b7374b1b868f198a05e44821ec2f7642351f221868d72ab385cefdb19ba1e1917b315085b2be422426ead40f57dd477545f76d540ba72e33b40dcc282940ca868af5c2cd40d617ca2d6b87298c491080d0dd0435647f4adbe8e5ec2025e3eb788d76e9928753b2de0ce67667c39d9ef7f6c9fb6f7e4e3bd2256889fb4974f54e9d43d95b36bd0a38d0a3729b2aedc829177d3acb76a4ef62bdfb0dd1e9f0e7fa994cc89c9dc216fa1f333034714bb3b40cd09c301aacd3635f8cd9f09fc331b649fd5c0cc2f0ab70b5a0883479b17ef34648eec3d90fa081278c9c66078a1c0fa680f1abaca12450130c50b496aecbf57a718d670c38a95879a88dbac7c86ec17e61f27aac4402fe058232c89fbc9b9c8ea00b91b379ee1888f0ee4ba6f31a7408136c7e73330eb8bda640e37814ec1d26eda7991992bcfc279d808a9d818210d52e27184a0e9232b73be12d90aa70a953f1908b21f28a896512d777785fae63a407bc6e33cdda44d56f0f9d7e4db7bea2fea685cd4169b4bdc7a382334c60a62fe5764b50c49b1f3e146d7fb859eaf4f920b9e4b"
	// // s := "0000475f05e647df85872864a97228422f332eb44c2a1c02c26b31cc715637130cc20000000121e67317cbc4be2aeb00677ad6462778a8f52274b9d605df2591b23027a87dff0000000700000011145984ae0000000000000000000000010000000117bbde3f175cd8c7327ab1f62ed7f00a5477d12e44032e12"

	// // decodeUTXO(s)

	// decodeContainer()

	// //decodeContainer("11Atn7rC6LxAVZ3Dyc8xxB8Uz3LMxzos6TjbojRsav3zcJECB919hqakKzUPUTdTiynWUfpMsL2j1nEbcW4WWP1ZkwG48R6DBvUX4s6MVsCm9CdtpU1WbpwSBiPrMKEsZaZkuheaqJdMMtCeyZ8EyFYvJNLMtwyaS43LTb")

	// dc := "11TGBQa1vccz69XSpeG8SmYdms1Q6HD5wHbRtr5vbPN8qHPSKtvXT9dFyJrMo4gwXNNCrwciqVSjMCUp7m5cn72Rjy1Gt5qZNWfGrqG2KcggmuNtFNGfKV5AdXtF3nCHw2UZnagQRNmxUNmJEbRtuozscrvBPc4AgTbh7p"
	// //dc := "11TGBQa1vccz69XSpeG8SmYdms1Q6HD5wHbRtr5vbPN8qHPSKtvqhuBTbQTzycBJf1VoEcokqzUD9WXck3AvnE5tsab5w4SKogXLtvdHF88xxcUKAfEqZfbo5ZjBSL6TWaN4NasrVyobeXoZXGABQJRXwVqXgaicEBkFsD"
	// //dc := "11Atn7rC6LxAVZ3Dyc8xxB8Uz3LMxzos6TjbojRsav3zcJECB919hqakKzUPUTdTiynWUfpMsL2j1nEbcW4WWP1ZkwG48R6DBvUX4s6MVsCm9CdtpU1WbpwSBiPrMKEsZaZkuheaqJdMMtCeyZ8EyFYvJNLMtwyaS43LTb"
	// dcUTXO(dc)

	makeResearchCSV("26HeCVKx9viUPp5EdkitRMB2tkdzF62dNW1snx9sZGe461m4V2", 100)

}
