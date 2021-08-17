package main

import (
	"Alexios/skiplist"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	size := 10
	list := skiplist.SkiplistInit(skiplist.String)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hash := make(map[string]int)
	for i := 0; i < size; i++ {
		bytes := make([]byte, 16)
		for j := 0; j < 16; j++ {
			b := r.Intn(26) + 65
			bytes[j] = byte(b)
		}
		hash[string(bytes[:])] = i
	}
	start := time.Now() // 获取当前时间
	for k, v := range hash {
		list.Insert(k, v)
	}
	elapsed := float64(time.Since(start).Seconds())
	fmt.Printf("插入数据%d条,耗时%f", size, elapsed)
	return
	count := 0
	list.Insert("1212121", 111)

	for k, v := range hash {
		res, err := list.Find(k)
		if err != nil {
			fmt.Println(err)
		} else {
			if res.Value() == v {
				count++
			}
		}
	}

	fmt.Println(count)
}
