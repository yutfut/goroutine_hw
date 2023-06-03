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

func SingleHash(data string, c chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	//fmt.Printf("data %s\n", data)
	md5Data := DataSignerMd5(data)
	//fmt.Printf("md5(data) %s\n", md5Data)
	crc32Md5Data := DataSignerCrc32(md5Data)
	//fmt.Printf("crc32(md5(data)) %s\n", crc32Md5Data)
	crc32Data := DataSignerCrc32(data)
	//fmt.Printf("crc32(data) %s\n", crc32Data)
	result := crc32Data + "~" + crc32Md5Data
	//fmt.Printf("result %s\n", result)
	c <- result
}

func MultiMultiHash(data string, i int, wg *sync.WaitGroup, c chan string) {
	defer wg.Done()
	result := DataSignerCrc32(strconv.Itoa(i) + data)
	//fmt.Printf("%d %s\n", i, result)
	c <- result
}

func MultiHash(c chan string, wg *sync.WaitGroup, hashArray *[]string) {
	defer wg.Done()
	a := <-c
	//fmt.Println(a)

	chanArray := make([]chan string, 0)

	for i := 0; i < 6; i++ {
		buffChan := make(chan string)
		chanArray = append(chanArray, buffChan)
		wg.Add(1)
		go MultiMultiHash(a, i, wg, buffChan)
	}

	finishString := <-chanArray[0]
	finishString += <-chanArray[1]
	finishString += <-chanArray[2]
	finishString += <-chanArray[3]
	finishString += <-chanArray[4]
	finishString += <-chanArray[5]
	//fmt.Printf("%s\n", finishString)
	*hashArray = append(*hashArray, finishString)
}

func main() {
	start := time.Now()
	wg := &sync.WaitGroup{}
	hashArray := make([]string, 0)
	for i := 0; i < 7; i++ {
		c1 := make(chan string)
		wg.Add(1)
		go SingleHash(strconv.Itoa(i), c1, wg)
		wg.Add(1)
		go MultiHash(c1, wg, &hashArray)
	}
	wg.Wait()

	fmt.Printf("--//--\n")
	sort.Strings(hashArray)

	result := hashArray[0]
	for j, hash := range hashArray {
		if j == 0 {
			continue
		}
		result += "_"
		result += hash
	}
	fmt.Println(result)
	fmt.Println(time.Now().Sub(start))
}
