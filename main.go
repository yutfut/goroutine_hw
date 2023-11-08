package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	dataSignerOverheat uint32 = 0
	DataSignerSalt            = ""
)

func OverheatLock() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
			fmt.Println("OverheatLock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

func OverheatUnlock() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
			fmt.Println("OverheatUnlock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

func DataSignerMd5(data string) string {
	OverheatLock()
	defer OverheatUnlock()
	data += DataSignerSalt
	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	time.Sleep(10 * time.Millisecond)
	return dataHash
}

func DataSignerCrc32(data string) string {
	data += DataSignerSalt
	crcH := crc32.ChecksumIEEE([]byte(data))
	dataHash := strconv.FormatUint(uint64(crcH), 10)
	time.Sleep(time.Second)
	return dataHash
}

func MD5(c chan string, data string) {
	c <- DataSignerMd5(data)
}

func CRC32(c chan string, data string) {
	c <- DataSignerCrc32(data)
}

func SingleHash(data string, sendToMulti chan string) {
	c1 := make(chan string)
	c2 := make(chan string)
	go MD5(c1, data)
	go CRC32(c2, data)
	go CRC32(c1, <-c1)

	sendToMulti <- <-c2 + "~" + <-c1
}

func MultiMultiHash(data string, i int, wg *sync.WaitGroup, result []string) {
	defer wg.Done()
	result[i] = DataSignerCrc32(strconv.Itoa(i) + data)
}

func MultiHash(receiveFromSingle chan string, sendToCombine chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	resultSingle := <-receiveFromSingle
	close(receiveFromSingle)

	resultArray := make([]string, 6)

	wgMulti := &sync.WaitGroup{}

	for i := 0; i < 6; i++ {
		wgMulti.Add(1)
		go MultiMultiHash(resultSingle, i, wgMulti, resultArray)
	}

	wgMulti.Wait()

	result := ""

	for _, item := range resultArray {
		result += item
	}

	fmt.Println(result)
	sendToCombine <- result
}

func CombineResults(receiveFromMilti chan string, returnResult chan string) {
	hashArray := make([]string, 0)

	for a := range receiveFromMilti {
		hashArray = append(hashArray, a)
	}

	sort.Strings(hashArray)

	result := hashArray[0]
	for j, hash := range hashArray {
		if j == 0 {
			continue
		}
		result += "_"
		result += hash
	}
	fmt.Println("-//-")
	returnResult <- result
	close(returnResult)
}

func main() {
	start := time.Now()
	wg := &sync.WaitGroup{}

	sendToCombine := make(chan string)
	receiveFromCombine := make(chan string)

	go CombineResults(sendToCombine, receiveFromCombine)

	for i := 0; i < 2; i++ {
		fromSingsToMulti := make(chan string)
		wg.Add(1)
		go SingleHash(strconv.Itoa(i), fromSingsToMulti)
		go MultiHash(fromSingsToMulti, sendToCombine, wg)
	}
	wg.Wait()
	close(sendToCombine)

	fmt.Println(<-receiveFromCombine)

	fmt.Println(time.Now().Sub(start))
}
