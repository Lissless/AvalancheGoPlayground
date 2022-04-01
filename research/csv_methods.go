package research

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"main.go/container"
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
