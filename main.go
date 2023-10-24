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
	c <- DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
	
}

func MultiMultiHash(data string, i int, wg *sync.WaitGroup, c chan string) {
	defer wg.Done()
	c <- DataSignerCrc32(strconv.Itoa(i) + data)
}

func MultiHash(c chan string, wg *sync.WaitGroup, hashArray *[]string) {
	defer wg.Done()
	a := <-c

	chanM := make(chan string)

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go MultiMultiHash(a, i, wg, chanM)
	}

	for i := 0; i < 6; i++ {
		*hashArray = append(*hashArray, <- chanM)
	}
	close(chanM)
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
