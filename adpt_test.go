package main

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"testing"
	"time"
)

// === 用户定义的强类型业务 ===

type OrderRequest struct {
	OrderId string `dust:"0"`
	UserId  int    `dust:"1"`
}

type OrderResponse struct {
	Status string
	Code   int
}

// CreateOrder 业务函数
func CreateOrder(pipe chan *anything.Mission, req OrderRequest) OrderResponse {
	fmt.Printf(">>> 业务执行: 创建订单 ID=%s, 用户=%d\n", req.OrderId, req.UserId)

	// 模拟业务处理...

	// 【关键修复点 1】: 任务完成后，主动告诉调度器退出
	// 如果不发这个，调度器会一直空转等待下一个任务
	pipe <- &anything.Mission{Name: anything.ExitFunction}

	return OrderResponse{Status: "Success", Code: 200}
}

func LandingABC(name string, stop bool) bool {
	if name == "" {
		return stop
	} else if name == "start" {
		return true
	}
	return false
}

func TestM(t *testing.T) {
	w := &anything.Wind{}
	w.Init()

	// 1. 适配
	genericFunc := w.Adapt(CreateOrder)
	genericFunc1 := w.Adapt(LandingABC)
	// 2. 注册
	w.M.Store("Order.Create", genericFunc)
	w.M.Store("ABC", genericFunc1)

	// 3. 调度
	// Schedule 会返回任务流的唯一 ID (key)
	key := w.Schedule("ABC", []any{"start", false})

	// 【关键修复点 2】: 正确等待任务结束
	// 不要使用 select {}，而是监听 Wind 暴露的退出信号通道
	if doneChan, ok := w.E[key]; ok {
		// 阻塞等待，直到 CreateOrder 发送 ExitFunction 并且 Wind 处理完毕
		select {
		case <-doneChan:
			t.Log("任务流正常结束")
		case <-time.After(2 * time.Second):
			t.Error("测试超时: 任务没有在规定时间内结束")
		}
	} else {
		t.Error("无法获取任务退出通道")
	}
}
