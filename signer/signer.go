package main

import "fmt"

func SingleHash(data string) {
	fmt.Printf("data %s\n", data)
	md5Data := DataSignerMd5(data)
	fmt.Printf("md5(data) %s\n", md5Data)
	crc32Md5Data := DataSignerCrc32(md5Data)
	fmt.Printf("crc32(md5(data)) %s\n", crc32Md5Data)
	crc32Data := DataSignerCrc32(data)
	fmt.Printf("crc32(data) %s\n", crc32Data)
	result := crc32Data + "~" + crc32Md5Data
	fmt.Printf("result %s\n", result)
}

func main() {
	SingleHash(string(rune(0)))
}
