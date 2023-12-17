package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func MD5(out chan string, data string) {
	out <- DataSignerMd5(data)
}

func CRC32(out chan string, data string) {
	out <- DataSignerCrc32(data)
}

func SingleHash(in chan string, out chan string) {
	c1 := make(chan string)
	c2 := make(chan string)
	go MD5(c1, <-in)
	go CRC32(c2, <-in)
	go CRC32(c1, <-c1)

	out <- <-c2 + "~" + <-c1
}

func ForMultiHash(th int, data string, result *string, wg *sync.WaitGroup) {
	defer wg.Done()
	*result = DataSignerCrc32(strconv.Itoa(th) + data)
}

func MultiHash(in chan string, out chan string) {
	data := <-in

	result := make([]string, 6)

	wg := &sync.WaitGroup{}

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go ForMultiHash(i, data, &result[i], wg)
	}

	wg.Wait()

	out <- strings.Join(result, "")
}

func CombineResults(in chan string, out chan string) {
	hashArray := make([]string, 0)

	for a := range in {
		hashArray = append(hashArray, a)
	}

	sort.Strings(hashArray)

	out <- strings.Join(hashArray, "_")
}

func ExecutePipeline(jobs ...job) {

}

func ExecutePipelineMock() {
	start := time.Now()

	c3 := make(chan string)
	c4 := make(chan string)

	go CombineResults(c3, c4)

	for i := 0; i < 2; i++ {
		c1 := make(chan string)
		c2 := make(chan string)
		c1 <- strconv.Itoa(i)
		go SingleHash(c1, c2)
		go MultiHash(c2, c3)
	}

	fmt.Println(<-c4)

	fmt.Println(time.Now().Sub(start))
}
