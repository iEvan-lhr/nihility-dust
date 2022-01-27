package main

import (
	"fmt"
	"nihility-dust/dust"
	"runtime"
	"strconv"
	"time"
)

func main() {
	nums := make([]any, 0)
	appends := make([]any, 0)
	for i := 0; i < 100000; i++ {
		appends = append(appends, i)
	}
	d := dust.Dust{
		Describe: "NumsTest",
		Counter:  runtime.NumCPU() * 1000,
	}
	start := time.Now()
	nums = d.AppendSlice(nums, appends)
	fmt.Println("Dust Done Time:", time.Now().Sub(start))
	fmt.Println("Dust lens:", len(nums))
	nums = nil
	start = time.Now()
	nums = append(nums, appends...)
	fmt.Println("Golang Done Time:", time.Now().Sub(start))
	fmt.Println("Golang lens:", len(nums))
	type TestStruct struct {
		id   int
		name string
	}

	str := make([]any, 0)
	appStr := make([]any, 0)
	for i := 0; i < 100000; i++ {
		appStr = append(appStr, TestStruct{
			id:   i,
			name: "Struct:" + strconv.Itoa(i),
		})
	}
	s := dust.Dust{
		Describe: "StrTest",
		Counter:  runtime.NumCPU() * 1000,
	}
	str = s.AppendSlice(str, appStr)
	fmt.Println("Dust Done Time:", time.Now().Sub(start))
	fmt.Println("Dust lens:", len(nums))
	nums = nil
	start = time.Now()
	nums = append(nums, appends...)
	fmt.Println("Golang Done Time:", time.Now().Sub(start))
	fmt.Println("Golang lens:", len(nums))
}
