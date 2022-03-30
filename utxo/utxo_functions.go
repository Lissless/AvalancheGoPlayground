package utxo

import (
	"context"
	"encoding/binary"
	"math/big"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

type UTXO struct {
	codecID     uint16
	txID        string
	outputIndex int
	assetID     string
	output      Output
}

type Output struct {
	typeID       int
	amount       big.Int
	lockTime     big.Int
	threshold    int
	no_addresses int
	addresses    []string
}

func GetRewardUTXOs(txID string) ([][]byte, error) {
	requester := rpc.NewEndpointRequester("https://api.avax.network", "/ext/P", "platform")
	ctx := context.Background()
	resp := new(platformvm.GetRewardUTXOsReply)
	ID, err := ids.FromString(txID)
	if err != nil {
		return nil, err
	}
	args := api.GetTxArgs{
		TxID:     ID,
		Encoding: formatting.CB58,
	}
	requester.SendRequest(ctx, "getRewardUTXOs", args, resp)
	// from Avalanche Go
	utxos := make([][]byte, len(resp.UTXOs))
	for i, utxoStr := range resp.UTXOs {
		utxoBytes, err := formatting.Decode(resp.Encoding, utxoStr)
		if err != nil {
			return nil, err
		}
		utxos[i] = utxoBytes
	}
	return utxos, err

}

func makeAddressArray(addressBytes []byte) ([]string, error) {
	numAddresses := len(addressBytes) / 20
	addrStringArr := make([]string, numAddresses)
	for i := 0; i < numAddresses; i++ {
		da, err := formatting.FormatAddress("P", constants.GetHRP(1), addressBytes[(i*20):((i+1)*20)])
		if err != nil {
			return nil, err
		}
		addrStringArr[i] = da
	}

	return addrStringArr, nil
}

func FormatUTXO(utxoBytes []byte) (*UTXO, error) {
	utxo := new(UTXO)

	codecIDBytes := utxoBytes[:2]
	transactionIDBytes := utxoBytes[2:34]
	outputIndexBytes := utxoBytes[34:38]
	assetIDBytes := utxoBytes[38:70]
	utxoOutputBytes := utxoBytes[70:]

	typeIDBytes := utxoOutputBytes[:4]
	amountBytes := utxoOutputBytes[4:12]
	locktimeBytes := utxoOutputBytes[12:20]
	thresholdBytes := utxoOutputBytes[20:24]
	addressAmountBytes := utxoOutputBytes[24:28]
	addressesBytes := utxoOutputBytes[28:]

	// make codecID
	codecID := binary.LittleEndian.Uint16(codecIDBytes)
	utxo.codecID = codecID

	// make transactionID
	transactionID, err := ids.ToID(transactionIDBytes)
	if err != nil {
		return nil, err
	}
	utxo.txID = transactionID.String()

	// make outputIndex
	outputIndex := int(binary.BigEndian.Uint32(outputIndexBytes))
	utxo.outputIndex = outputIndex

	// make assetID
	assetID, err := ids.ToID(assetIDBytes)
	if err != nil {
		return nil, err
	}
	utxo.assetID = assetID.String()

	// make TypeID
	typeID := int(binary.BigEndian.Uint32(typeIDBytes))
	utxo.output.typeID = typeID

	// make Amount
	amount := new(big.Int)
	amount.SetBytes(amountBytes)
	utxo.output.amount = *amount

	// make Locktime
	locktime := new(big.Int)
	locktime.SetBytes(locktimeBytes)
	utxo.output.lockTime = *locktime

	// make Threshold
	threshold := int(binary.BigEndian.Uint32(thresholdBytes))
	utxo.output.threshold = threshold

	// make Address Amount
	addressAmount := int(binary.BigEndian.Uint32(addressAmountBytes))
	utxo.output.no_addresses = addressAmount

	// make Addresses
	addresses, err := makeAddressArray(addressesBytes)
	if err != nil {
		return nil, err
	}
	utxo.output.addresses = addresses

	return utxo, nil

}
