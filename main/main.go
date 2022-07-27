package main

import (
	"fmt"
	"nihility-dust/anything"
	"nihility-dust/dust"
	"nihility-dust/test"
	//"github.com/iEvan-lhr/nihility-dust/wind"
	"math/rand"
	"time"
)

func main() {

	w := anything.Wind{}
	//方法注册
	w.Register(&dust.Dust{}, &test.Ran{})
	//执行器初始化
	w.Init()
	rand.Seed(time.Now().UnixNano())
	//入口
	start := time.Now()
	//key := w.Schedule("CheckIsBig", 25)
	// 出口
	mis := <-anything.DoChanN("CheckIsBig", []any{25})
	//mission, _ := w.A.Load(key)
	fmt.Println(mis.Pursuit)

	fmt.Println(time.Now().Sub(start))
}
