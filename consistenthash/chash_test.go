package consistenthash

import (
	"fmt"
	"strconv"
	"testing"
)

func TestMap_Add(t *testing.T) {
	chash := New(3, func(data []byte) uint32 {
		atoi, _ := strconv.Atoi(string(data))
		return uint32(atoi)
	})

	// 02/12/22   04/14/24   06/16/26
	// 2 4 6 12 14 16 22 26
	chash.Add("2", "4", "6")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		get := chash.Get(k)
		fmt.Println(v, get)
	}

	chash.Add("8")

	//"27":"2"  ==> "27":"8"
	get := chash.Get("27")
	fmt.Println("27", get)

	get = chash.Get("127")
	fmt.Println("127=2?", get) // 2
}
