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
	"sync"
)

type transaction struct {
	id     string
	inputs []string
}

type output struct {
	id    string
	count int
}

var wg sync.WaitGroup
var lock sync.Mutex

const (
	PAGE_SIZE = 25

	BLOCK_HEIGHT_URL = "https://blockstream.info/api/block-height/"

	BLOCK_TRNASACTIONS_URL = "https://blockstream.info/api/block/%s/txs/%s"

	BLOCK_TOTAL_TRANSACTIONS = "https://blockstream.info/api/block/%s/txids"
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
	totalTransactionMap := make(map[string]transaction)
	hash, err := fetchDataFromUrl(BLOCK_HEIGHT_URL + block_height)
	if err != nil {
		return nil, err
	}
	if hash == "Block not found" {
		return nil, errors.New("Block not found")
	}
	fmt.Println("hashrespmap: ", hash)

	txidsresp, er := fetchDataFromUrl(fmt.Sprintf(BLOCK_TOTAL_TRANSACTIONS, hash))
	if er != nil {
		return nil, er
	}

	var totaltxs []interface{}
	er = json.Unmarshal([]byte(txidsresp), &totaltxs)
	if er != nil {
		return nil, er
	}
	page_index := 0
	for (page_index * PAGE_SIZE) < len(totaltxs) {
		wg.Add(1)
		go updateTransactionsFromIndex(hash, page_index, totalTransactionMap)
		page_index++
	}
	wg.Wait()
	fmt.Println("total transactions: ", totalTransactionMap)
	fmt.Println("transaction map: ", getCountMapForTrasactions(totalTransactionMap))
	return getTop10(getCountMapForTrasactions(totalTransactionMap)), nil
}

func updateTransactionsFromIndex(hash string, idx int, tMap map[string]transaction) {
	respbody, e := fetchDataFromUrl(fmt.Sprintf(BLOCK_TRNASACTIONS_URL, hash, strconv.Itoa(idx*PAGE_SIZE)))
	if e != nil {
		fmt.Println("error fetching transactions from index: ", idx, e)
		return
	}
	var transactionsdata []interface{}
	err := json.Unmarshal([]byte(respbody), &transactionsdata)
	if err != nil {
		fmt.Println("error fetching transactions from index: ", idx, err)
		return
	}
	lock.Lock()
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
		tMap[tran.id] = tran
	}
	lock.Unlock()
	wg.Done()
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

func getCountMapForTrasactions(tMap map[string]transaction) map[string]int {
	transactioncountmap := make(map[string]int)

	for _, tn := range tMap {
		transactioncountmap[tn.id] = findAncestorsForTransaction(tn, tMap)
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
