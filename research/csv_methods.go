package research

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/Lissless/AvalancheGoPlayground/container"
)

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

func checkNotEmpty(block map[string]interface{}) bool {
	for k := range block {
		if k == "tx" {
			return true
		}
	}
	return false
}

func CheckBlockIDFromIndexAccuracy(startIndex int, numEntries int) {
	numSuccess := 0
	for i := 0; i < numEntries; i++ {
		time.Sleep((1 * time.Second))
		fmt.Printf("Processing entry %d of %d \n", (i + 1), numEntries)
		blockID, success := container.GetBlockIDFromIndex(float64(startIndex + i))
		resp := container.GetBlock(blockID)
		if success && resp.Block != nil {
			numSuccess++
		}
	}
	fmt.Printf("GetBlockIDFromIndex is %d %% Accurate", ((numSuccess / numEntries) * 100))
}

func MakeBlockIDAtCSV(blockID string, numEntries int) {
	csvFile, err := os.Create("containerResearchBlockIDAt.csv")
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
		fmt.Println("Processing entry: ", (i + 1))
		chain = "N/A"
		chainName = "N/A"
		expinp = "N/A"
		resp := container.GetBlock(blockID)
		block := resp.Block.(map[string]interface{})
		if checkNotEmpty(block) {
			chain, chainName = getChainType(block)
			expinp = getExportOrImportType(block)
		}

		height := block["height"].(float64)
		containerID, cont := container.DecodeContainer(height)
		possibleIDs, appearedAt, transactionID := container.FindBlockID(cont.Bytes, blockID)
		row := []string{blockID, fmt.Sprintf("%f", height), containerID, strconv.Itoa(possibleIDs), strconv.Itoa(appearedAt), transactionID, chain, chainName, expinp}

		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}

		blockID = block["parentID"].(string)
	}
}

func MakeFindRelevantIDsCSV(blockID string, numEntries int) {
	csvFile, err := os.Create("containerResearchAllRelevantBlockIDs.csv")
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
		resp := container.GetBlock(blockID)
		block := resp.Block.(map[string]interface{})
		height := block["height"].(float64)
		containerID, cont := container.DecodeContainer(height)
		fmt.Println("Doing entry ", (i + 1), " of ", numEntries)
		valid := container.FindAllValidBlockIDs(cont.Bytes)
		row := []string{fmt.Sprintf("%f", height), containerID, valid}
		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}

		blockID = block["parentID"].(string)
	}

}

func MakeFindRelevantTransIDsCSV(blockID string, numEntries int) {
	csvFile, err := os.Create("containerResearchAllRelevantTransIDs.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	initialRow := []string{"Block Height", "Container ID", "Found Valid Transaction IDs"}
	if err := csvwriter.Write(initialRow); err != nil {
		log.Fatalln("error writing record to file", err)
	}

	for i := 0; i < numEntries; i++ {
		resp := container.GetBlock(blockID)
		block := resp.Block.(map[string]interface{})
		height := block["height"].(float64)
		containerID, cont := container.DecodeContainer(height)
		fmt.Println("Doing entry ", (i + 1), " of ", numEntries)
		valid := container.FindAllValidTransactionIDs(cont.Bytes)
		row := []string{fmt.Sprintf("%f", height), containerID, valid}
		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}

		blockID = block["parentID"].(string)
	}

}

func PrintBlockTxType(index float64, numEntries int, intervals int, stop int) {
	csvFile, err := os.Create("blockResearchTxType2.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	initialRow := []string{"index Start", "Index End", "Has Tx Data", "Has Tx ID", "Empty", "RelevantBlkID"}
	if err := csvwriter.Write(initialRow); err != nil {
		log.Fatalln("error writing record to file", err)
	}
	index = index - float64(intervals)

	for int(index) < stop {
		fmt.Printf("Looking at entries %f to %f, stopping at %d", index, index+float64(intervals), stop)

		index = index + float64(intervals)
		blockID, _ := container.GetBlockIDFromIndex(index)

		txIDPresent := 0
		txIDNotPresent := 0
		empty := 0
		for i := 0; i < numEntries; i++ {
			time.Sleep((1 * time.Second))
			fmt.Println("Processing Entry: ", (i + 1))
			resp := container.GetBlock(blockID)
			block := resp.Block.(map[string]interface{})
			if checkNotEmpty(block) {
				tx := block["tx"].(map[string]interface{})
				utx := tx["unsignedTx"].(map[string]interface{})
				for k := range utx {
					if k == "txID" {
						row := []string{"N/A", "N/A", "N/A", "N/A", "N/A", blockID}
						if err := csvwriter.Write(row); err != nil {
							log.Fatalln("error writing record to file", err)
						}
						txIDPresent++
						break
					}
					txIDNotPresent++
					break
				}
			} else {
				empty++
			}
			blockID = block["parentID"].(string)
		}

		row := []string{fmt.Sprintf("%f", index), fmt.Sprintf("%d", int(index)+numEntries), fmt.Sprintf("%d", txIDPresent), fmt.Sprintf("%d", txIDNotPresent), fmt.Sprintf("%d", empty), "N/A"}
		if err := csvwriter.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	// fmt.Printf("From Entries %d to %d\n", int(index), (int(index) - numEntries))
	// fmt.Printf("Out of %d:\n%d had transaction data \n%d had transactionID \n%d were empty", numEntries, txIDPresent, txIDNotPresent, empty)

}

func FindTxAbove(initialTxIndex int, stop int) {
	blockID, _ := container.GetBlockIDFromIndex(float64(initialTxIndex))
	resp := container.GetBlock(blockID)
	block := resp.Block.(map[string]interface{})
	IDFlag := false
	if checkNotEmpty(block) {
		Itx := block["tx"].(map[string]interface{})
		utx := Itx["unsignedTx"].(map[string]interface{})
		for k := range utx {
			if k == "txID" {
				IDFlag = true
				break
			}
		}
		if IDFlag {
			fmt.Println("This starting index is invalid: has Transaction ID")
		}

		for i := initialTxIndex; i < stop; i++ {
			fmt.Println("checking transaction at index: ", i)
			time.Sleep((1 * time.Second))
			blockID, _ := container.GetBlockIDFromIndex(float64(i))
			resp := container.GetBlock(blockID)
			block := resp.Block.(map[string]interface{})
			if checkNotEmpty(block) {
				tx := block["tx"].(map[string]interface{})
				utx := tx["unsignedTx"].(map[string]interface{})
				for k := range utx {
					if k == "txID" {
						resp := container.GetTransaction(utx["txID"].(string))
						if resp.Tx != nil {
							FTx := resp.Tx.(map[string]interface{})
							fmt.Println("Transaction found at index: ", i, " BlockID: ", blockID)
							if reflect.DeepEqual(Itx, FTx) {
								fmt.Println("Found at index: ", i)
							}
						}
						break
					}
					break
				}
			}

		}
		fmt.Println("Not Found")

	} else {
		fmt.Println("This starting index is invalid: Does not contain a transaction")
	}

}
