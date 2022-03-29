package container

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/utils/formatting"
	avajson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/utils/rpc"
)

func DecodeContainer(height float64) (string, indexer.Container) {

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

func MakeContainerTypeID(contData []byte) string {
	transactionSlice := contData[:6]
	transactionID := ""
	for i := 0; i < 6; i++ {
		subject := transactionSlice[i]
		numVal := int(subject)
		transactionID = transactionID + strconv.Itoa(numVal)
	}
	return transactionID
}

func GetBlock(blockID string) api.GetBlockResponse {
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

func FindAllValidBlockIDs(containerData []byte) string {
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
		resp := GetBlock(stringID)
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

func investigate(transactionID string, contData []byte) string {
	fmt.Println(`Block ID not found for ID "` + transactionID + `", Looking for BlockID Location`)
	blockID := FindAllValidBlockIDs(contData)
	blockID = blockID[:(len(blockID) - 1)]
	_, appearedAt, _ := FindBlockID(contData, blockID)
	fmt.Println("Block ID: ", blockID, " Found at Index: ", appearedAt)
	return fmt.Sprintf("Block ID: %s Found at Index: %d", blockID, appearedAt)
}

func FindBlockID(containerData []byte, wanted string) (int, int, string) {
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

	transactionID := MakeContainerTypeID(containerData)

	return (length - 31), found, transactionID
}

func GetBlockIDFromIndex(index float64) (string, bool) {
	if index < 0 {
		return "Invalid Index", false
	}
	_, container := DecodeContainer(index)
	return GetBlockIDFromContainer(container.Bytes)
}

func GetBlockIDFromContainer(contData []byte) (string, bool) {
	transactionID := MakeContainerTypeID(contData)

	var slice []byte
	switch transactionID {
	case "000000":
		length := len(contData)
		check := []int{6, 68, 1253, 1321, 1344}
		found := false
		var blockID string
		var oof error
		for _, val := range check {
			if (val + 32) < length {
				slice = contData[val:(val + 32)]
				blockID, oof = formatting.EncodeWithChecksum(formatting.CB58, slice)
				if oof != nil {
					return "Encoding Unsuccessful", false
				}
				resp := GetBlock(blockID)
				if resp.Block != nil {
					found = true
					break
				}
			}
		}
		if !found {
			investigate(transactionID, contData)
		}
		return blockID, true
	case "000001":
		slice = contData[48:80]
	case "000002":
		slice = contData[6:38]
	case "000004":
		slice = contData[6:38]
	default:
		return fmt.Sprintf("Unknown Identifier: %s, %s", transactionID, investigate(transactionID, contData)), false
	}
	blockID, err := formatting.EncodeWithChecksum(formatting.CB58, slice)
	if err != nil {
		return "Encoding Unsuccessful", false
	}

	return blockID, true

}
