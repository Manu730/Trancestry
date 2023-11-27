package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
)

type transaction struct {
	id     string
	inputs []string
}

type output struct {
	id    string
	count int
}

const (
	PAGE_SIZE = 25

	BLOCK_HEIGHT_URL = "https://blockstream.info/api/block-height/"

	BLOCK_TRNASACTIONS_URL = "https://blockstream.info/api/block/%s/txs/%s"
)

func main() {
	if len(os.Args) == 1 {
		panic("block info not provided")
	}
	block := os.Args[1]
	fmt.Println("fetching transaction map for block: ", block)
	o, err := getTop10ForBlock(block)
	if err != nil {
		fmt.Println("error fetching transaction map ", err)
		panic(err)
	}
	fmt.Println("top10 transactions for block: ", o)
}

func getTop10ForBlock(block_height string) ([]output, error) {
	var totalTransactions []transaction
	hash, err := fetchDataFromUrl(BLOCK_HEIGHT_URL + block_height)
	if err != nil {
		return nil, err
	}
	if hash == "Block not found" {
		return nil, errors.New("Block not found")
	}
	fmt.Println("hashrespmap: ", hash)
	page_num := 0
	for {
		respbody, e := fetchDataFromUrl(fmt.Sprintf(BLOCK_TRNASACTIONS_URL, hash, strconv.Itoa(page_num*PAGE_SIZE)))
		if e != nil {
			return nil, e
		}
		if respbody == "start index out of range" {
			break
		}
		var transactionsdata []interface{}
		err := json.Unmarshal([]byte(respbody), &transactionsdata)
		if err != nil {
			return nil, err
		}

		for _, trans := range transactionsdata {
			tranmap := trans.(map[string]interface{})
			var tran transaction
			tran.id = tranmap["txid"].(string)
			inputs := reflect.ValueOf(tranmap["vin"])
			for i := 0; i < inputs.Len(); i++ {
				iput := inputs.Index(i).Interface()
				inputmap := iput.(map[string]interface{})
				inputid := inputmap["txid"].(string)
				tran.inputs = append(tran.inputs, inputid)
			}
			totalTransactions = append(totalTransactions, tran)
		}
		page_num++
	}
	fmt.Println("total transactions: ", totalTransactions)
	fmt.Println("transaction map: ", getCountMapForTrasactions(totalTransactions))
	return getTop10(getCountMapForTrasactions(totalTransactions)), nil
}

func getTop10(m map[string]int) []output {
	var outs []output
	var sets []int
	for _, v := range m {
		sets = append(sets, v)
	}
	sort.Ints(sets)
	for i := len(sets) - 1; i >= 0; i-- {
		for k, v := range m {
			if v == sets[i] {
				o := new(output)
				o.id = k
				o.count = v
				outs = append(outs, *o)
			}
		}
	}
	return outs
}

func getCountMapForTrasactions(totaltrans []transaction) map[string]int {
	transactioncountmap := make(map[string]int)
	validTransactionMap := make(map[string]transaction)
	for _, t := range totaltrans {
		validTransactionMap[t.id] = t
	}

	for _, tn := range totaltrans {
		transactioncountmap[tn.id] = findAncestorsForTransaction(tn, validTransactionMap)
	}
	return transactioncountmap
}

func findAncestorsForTransaction(tran transaction, validmap map[string]transaction) int {
	if isAllInputsInvalid(tran.inputs, validmap) {
		return 0
	}
	ancestorcount := 0
	for _, in := range tran.inputs {
		if v, ok := validmap[in]; ok {
			ancestorcount += 1 + findAncestorsForTransaction(v, validmap)
		}
	}
	return ancestorcount
}

func isAllInputsInvalid(inputs []string, validmap map[string]transaction) bool {
	for _, in := range inputs {
		if _, ok := validmap[in]; ok {
			return false
		}
	}
	return true
}

func fetchDataFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error fetching data from url: ", url)
		return "", err
	}
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		fmt.Println("error reading data: ", e)
		return "", e
	}
	return string(body), nil
}
