package main

import (
	"sync"
	"strconv"
	"sort"
	"fmt"
)

func MD5(c chan string, data string) {
	c <- DataSignerMd5(data)
}

func CRC32(c chan string, data string) {
	c <- DataSignerCrc32(data)
}

// func SingleHash(data string, sendToMulti chan string) {
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

func ExecutePipeline(input chan string, output chan string) {
	sendToCombine := make(chan string)
	receiveFromCombine := make(chan string)
	go CombineResults(sendToCombine, receiveFromCombine)
	fromSingsToMulti := make(chan string)
	go SingleHash(<-input, fromSingsToMulti)
	wg := &sync.WaitGroup{}
	go MultiHash(fromSingsToMulti, sendToCombine, wg)
	wg.Wait()
	close(sendToCombine)
	fmt.Println(<-receiveFromCombine)
}
