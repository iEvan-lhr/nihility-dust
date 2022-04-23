package main

import (
	"fmt"
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
	key := w.Schedule("CheckIsBig", 25)
	// 出口
	<-w.E[key]
	mission, _ := w.A.Load(key)
	fmt.Println(mission.([]any)[0])

	fmt.Println(time.Now().Sub(start))
}
