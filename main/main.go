package main

import (
	"fmt"
	"time"
)

func main() {

	//w := anything.Wind{}
	////方法注册
	//w.Register(&dust.Dust{}, &test.Ran{})
	////执行器初始化
	//w.Init()
	//rand.Seed(time.Now().UnixNano())
	////入口
	start := time.Now()
	fmt.Println("Test")
	////key := w.Schedule("CheckIsBig", 25)
	//// 出口
	//mis := <-anything.DoChanN("CheckIsBig", []any{25})
	////mission, _ := w.A.Load(key)
	//fmt.Println(mis.Pursuit)
	//of := reflect.ValueOf(&name{})
	//log.Println(of.Kind() == 22)
	//of1 := reflect.ValueOf(testing)
	//pc := runtime.FuncForPC(of1.Pointer())
	//log.Println(pc.Name())
	//log.Println(of1.Call([]reflect.Value{reflect.ValueOf("ssss")}))
	fmt.Println(time.Now().Sub(start))
}
