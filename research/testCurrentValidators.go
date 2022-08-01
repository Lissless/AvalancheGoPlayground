package research

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

func GetCurrentValidators(rounds int) bool {
	requester := rpc.NewEndpointRequester("https://api.avax.network", "/ext/P", "platform")
	ctx := context.Background()

	// resp := deepCompare("dataOdd.txt", "dataEven.txt")
	// fmt.Println(resp)

	// test1 := "test"
	// test2 := "t3st"
	// dmp := diffmatchpatch.New()
	// diffs := dmp.DiffMain(test1, test2, false)
	// fmt.Println(dmp.DiffPrettyText(diffs))

	different := 0

	// var even string
	var evenArr []string
	// var odd string
	var oddArr []string

	var evenFile *os.File
	var oddFile *os.File
	var err error

	for j := 0; j < rounds; j++ {

		if j%2 == 0 {
			evenFile, err = os.Create("dataEven.txt")
			if err != nil {
				return false
			}
		} else {
			oddFile, err = os.Create("dataOdd.txt")
			if err != nil {
				return false
			}
		}

		fmt.Printf("Evaluating round: %d \n", (j + 1))

		out := new(platformvm.GetCurrentValidatorsReply)
		args := &platformvm.GetCurrentValidatorsArgs{}
		requester.SendRequest(ctx, "getCurrentValidators", args, out)

		for _, val := range out.Validators {
			data := val.(map[string]interface{})

			rewardOwner := data["rewardOwner"].(map[string]interface{})
			addresses := rewardOwner["addresses"].([]interface{})
			addr := addresses[0].(string)

			if j%2 == 0 {
				evenArr = append(evenArr, addr)
			} else {
				oddArr = append(oddArr, addr)
			}

		}

		if j%2 == 0 {
			sort.Strings(evenArr)
			for _, entry := range evenArr {
				evenFile.WriteString(entry + " \n")
			}
			evenFile.Close()
		} else {
			sort.Strings(oddArr)
			for _, entry := range oddArr {
				oddFile.WriteString(entry + " \n")
			}
			oddFile.Close()
		}

		if j != 0 {
			equal := deepCompare("dataEven.txt", "dataOdd.txt")
			if !equal {
				different++
			}
		}

		evenFile.Close()
		oddFile.Close()
		if j != 0 {
			if j%2 == 0 {
				e := os.Remove("dataOdd.txt")
				if e != nil {
					log.Fatal(e)
				}
			} else {
				e := os.Remove("dataEven.txt")
				if e != nil {
					log.Fatal(e)
				}
			}
		}

		if j != (rounds - 1) {
			time.Sleep((time.Minute * 29))
		}

	}

	fmt.Printf("Get Current Validators Was different: %d out of %d times", different, (rounds - 1))

	return true
}

func deepCompare(file1, file2 string) bool {
	sf, err := os.Open(file1)
	if err != nil {
		log.Fatal(err)
	}

	df, err := os.Open(file2)
	if err != nil {
		log.Fatal(err)
	}

	sscan := bufio.NewScanner(sf)
	dscan := bufio.NewScanner(df)

	for sscan.Scan() {
		dscan.Scan()
		if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
			return false
		}
	}

	sf.Close()
	df.Close()

	time.Sleep((time.Minute * 1))

	return true
}
