package main

import (
	"fmt"
	"nihility-dust/dust"
)

func main() {
	scanner:=dust.Scanner{}
	dust.AddDust("获取Http请求","GetHttp")
	fmt.Println(dust.Building("获取Http请求","http://","ssss"),scanner)
}


