package signer

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func MD5(out chan string, data string) {
	out <- DataSignerMd5(data)
}

func CRC32(out chan string, data string) {
	out <- DataSignerCrc32(data)
}

func SingleHash(in, out chan interface{}) {
	for i := range in {
		data := strconv.Itoa(i.(int))

		c1 := make(chan string)
		c2 := make(chan string)

		go MD5(c1, data)
		go CRC32(c2, data)
		go CRC32(c1, <-c1)

		go func() {
			out <- <-c2 + "~" + <-c1
		}()
	}
}

func ForMultiHash(th int, data string, result *string, wg *sync.WaitGroup) {
	defer wg.Done()
	*result = DataSignerCrc32(strconv.Itoa(th) + data)
}

func MultiHash(in, out chan interface{}) {
	for i := range in {
		data := i.(string)

		result := make([]string, 6)

		wg := &sync.WaitGroup{}

		for i := 0; i < 6; i++ {
			wg.Add(1)
			go ForMultiHash(i, data, &result[i], wg)
		}

		wg.Wait()

		out <- strings.Join(result, "")
	}
}

func CombineResults(in, out chan interface{}) {
	hashArray := make([]string, 0)

	for a := range in {
		hashArray = append(hashArray, a.(string))
	}

	sort.Strings(hashArray)
	out <- strings.Join(hashArray, "_")
}

func jobRunner(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer func() {
		close(out)
		wg.Done()
	}()

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
