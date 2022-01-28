package main

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/dust"
	"time"
)

func main() {
	start := time.Now()
	w := dust.Wind{}
	//方法注册
	w.Register(&dust.Dust{})
	//执行器初始化
	w.Init()
	//入口
	w.Schedule("StartMission", "dust/test.html")
	_, ok := w.A["StartMission"]
	for !ok {
		// 出口
		_, ok = w.A["StartMission"]
	}
	fmt.Println(time.Now().Sub(start))
}
