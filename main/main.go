package main

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"github.com/iEvan-lhr/nihility-dust/dust"
	"github.com/iEvan-lhr/nihility-dust/test"
	"github.com/iEvan-lhr/nihility-dust/wind"
	"math/rand"
	"time"
)

func main() {

	w := wind.Wind{}
	//方法注册
	w.Register(&dust.Dust{}, &test.Ran{})
	//执行器初始化
	w.Init()
	rand.Seed(time.Now().UnixNano())
	//入口
	start := time.Now()
	w.Schedule("CheckIsBig", "dust/test.html")
	w.Schedule("CheckIsBig", "dust/test.html")
	w.Schedule("CheckIsBig", "dust/test.html")
	w.Schedule("CheckIsBig", "dust/test.html")

	for {
		// 出口
		if mission, ok := w.A.Load("CheckIsBig" + anything.ExitFunction); ok {
			fmt.Println(mission)
			break
		}
	}
	fmt.Println(time.Now().Sub(start))
}
