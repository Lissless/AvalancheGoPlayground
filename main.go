package main

import (
	"fmt"

	"github.com/Lissless/AvalancheGoPlayground/container"
)

func main() {
	// blockID, success := container.GetBlockIDFromIndex(1355477)
	// if success {
	// 	fmt.Println("Block ID: ", blockID)
	// } else {
	// 	fmt.Println("Error: ", blockID)
	// }

	// utxoBytes, _ := utxo.GetRewardUTXOs("YS5hpR5Nk5oxUBsSiWpntPp2L76nwcogwgoE7e7FUb3xt263i")
	// utxo, _ := utxo.FormatUTXO(utxoBytes[0])
	// fmt.Println("UTXO: ", utxo)

	// _, data := container.DecodeContainer(392739)

	// //look := container.GetTransaction("EHcY7bWkzni1FoF7RkaVhoPujxwkP3Rnmk8bCgEaQa6BrKBbr")
	// res := container.FindAllValidTransactionIDs(data.Bytes)
	// fmt.Println("Results: ", res)

	// // research.MakeFindRelevantTransIDsCSV("26HeCVKx9viUPp5EdkitRMB2tkdzF62dNW1snx9sZGe461m4V2", 100)

	// fmt.Println("Done")

	// thing, _ := container.GetBlockIDFromIndex(150001)
	// fmt.Println(thing)

	// research.PrintBlockTxType(1200000, 50, 100000, 1400000)
	// for i := 0; i < 100; i++ {
	// 	time.Sleep((1 * time.Second))
	// 	bid, _ := container.GetBlockIDFromIndex(float64(1407001 + i))
	// 	fmt.Printf("Entry %d: %s\n", i, bid)
	// }

	// research.FindTxAbove(1407003, 1407303)
	// research.MakeBlockIDAtCSV("k9od5nnQEwZrA1D5ycopVqDiuuKyCMgvEPhtH4n1pq4f1Cr8z", 200)

	//index := container.GetIndex()
	// research.PrintBlockTxType()
	// fmt.Println(index)

	// research.GetCurrentValidators(2)

	ID, _ := container.HexTo58Converter("0x633F5500A87C3DbB9c15f4D41eD5A33DacaF4184")

	// ID, cont := container.DecodeContainer(1595565)
	fmt.Println(ID)
	// fmt.Println(container.FindBlockID(cont.Bytes, "Fc4o47iaXM4XYcwcBrKJp6VELUXv21WSUNgMesfspdpNgJQ8L"))

}
