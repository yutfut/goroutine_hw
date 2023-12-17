package signer

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

func SingleHash(in, out chan interface{}) {
	fmt.Println(1)

	data := (<- in).(string)

	c1 := make(chan string)
	c2 := make(chan string)
	go MD5(c1, data)
	go CRC32(c2, data)
	go CRC32(c1, <-c1)

	out <- <-c2 + "~" + <-c1
	fmt.Println(2)
}

func ForMultiHash(th int, data string, result *string, wg *sync.WaitGroup) {
	fmt.Println(5)
	defer wg.Done()
	*result = DataSignerCrc32(strconv.Itoa(th) + data)
	fmt.Println(6)
}

func MultiHash(in, out chan interface{}) {
	fmt.Println(3)

	data := (<- in).(string)

	result := make([]string, 6)

	wg := &sync.WaitGroup{}

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go ForMultiHash(i, data, &result[i], wg)
	}

	wg.Wait()

	out <- strings.Join(result, "")
	fmt.Println(4)
}

func CombineResults(in, out chan interface{}) {
	fmt.Println(7)
	hashArray := make([]string, 0)

	for a := range in {
		hashArray = append(hashArray, a.(string))
	}

	sort.Strings(hashArray)

	out <- strings.Join(hashArray, "_")
	fmt.Println(8)
}

func jobRunner(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer func () {
		close(out)
		wg.Done()
	} ()

	job(in, out)
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, job := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go jobRunner(job, in, out, wg)
		in = out
	}

	wg.Wait()
}

func ExecutePipelineMock() {
	start := time.Now()

	c3 := make(chan interface{})
	c4 := make(chan interface{})

	go CombineResults(c3, c4)

	go func() {
		time.Sleep(20 * time.Second) 
		close(c3)
	} ()

	for i := 0; i < 10; i++ {
		c1 := make(chan interface{})
		c2 := make(chan interface{})
		go SingleHash(c1, c2)
		go MultiHash(c2, c3)
		c1 <- strconv.Itoa(i)
	}

	result := <-c4
	fmt.Println(result.(string))

	fmt.Println(time.Now().Sub(start))
}
