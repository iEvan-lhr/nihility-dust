

# nihility-dust

> **“纵使世界上最微小的一部分，也能创造一切”** 
> 
> **Even the dust, if united, can create everything.**

`nihility-dust` 是一个基于 Go 1.18+ (泛型与 Channel) 构建的**轻量级协程调度与任务编排引擎**。

它旨在解决传统串行代码中复杂的依赖耦合问题，通过**无序执行**（Unordered Execution）和**动态调度**，让程序能够充分利用多核性能，实现任务的自动并行与流转。

## 📖 核心理念 (Philosophy)

在传统的编程模型中，我们往往需要按部就班地组织代码（如 A -> B -> C -> D），这不仅增加了代码的耦合度，也导致了 CPU 等待和性能浪费。

`nihility-dust` 采用了一种新的思路：

* **解耦 (Decoupling)**: 程序员无需显式管理复杂的函数调用链。
* **调度 (Scheduling)**: 由 `Wind` 调度器决定是否/何时调用方法。
* **流转 (Flow)**: 程序只需要从 Channel 中获取数据、执行操作、并将结果推入下一个环节。

通过这种方式，原本需要线性等待的任务（预期耗时 A+B+C+D）可以被优化为并行执行（预期耗时约等于 (A+B+C+D)/N），极大提升执行效率。

## ✨ 特性 (Features)

* **Wind 调度器**: 核心控制器，负责任务的分发、协程管理和生命周期控制。
* **Mission 驱动**: 基于 `Channel` 的任务流转机制，支持同步/异步任务调度。
* **Easy Model**: 提供简易模式 (`easyModel`)，支持快速注册和调用单体函数，便于开发和测试。
* **雪花算法集成**: 内置 `Worker` (基于 `anything/worker.go`)，提供高效的分布式唯一 ID 生成能力。
* **反射机制**: 支持动态方法注册与调用，极大地提高了代码的灵活性。

## 🛠️ 安装 (Installation)

```bash
go get github.com/iEvan-lhr/nihility-dust

```

*(注意：项目依赖 Go 1.18 或更高版本)*

## 🚀 快速开始 (Quick Start)

### 1. 定义任务 (Define a Mission)

任务函数需要遵循特定的签名：接收一个任务通道 `chan *anything.Mission` 和参数切片 `[]any`。

```go
package main

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
)

// 定义一个示例任务函数
func MyTask(mission chan *anything.Mission, data []any) {
	fmt.Println("Start MyTask with data:", data)
	
	// 处理业务逻辑...
	result := "Task Completed"

	// 如果有后续任务，可以将结果发送回 channel
	// 或者发送退出信号
	mission <- &anything.Mission{
        Name: anything.ExitFunction, // 结束当前任务流
        Pursuit: []any{result},
    }
}

```

### 2. 注册与运行 (Register & Run)

你可以使用 `EasyMission` 模式快速注册并运行任务。

```go
package main

import (
	"fmt"
	"time"
	"github.com/iEvan-lhr/nihility-dust/anything"
)

func main() {
	// 1. 注册任务
	// 支持注册结构体方法或普通函数
	anything.AddEasyMission([]any{MyTask})

	// 2. 启动任务
	// DoChanN 用于启动一个任务管道，返回一个 channel 用于接收结果
	// 参数: "MyTask" (任务名), []any{"InputData"} (输入参数)
	resultChan := anything.DoChanN("MyTask", []any{"Hello Dust"})

	// 3. 等待并获取结果
	select {
	case res := <-resultChan:
		fmt.Println("Received result:", res.Pursuit)
	case <-time.After(time.Second * 3):
		fmt.Println("Timeout")
	}
}

```

## 🏗️ 架构说明 (Architecture)

### 核心组件

* **anything 包**: 包含核心逻辑。
* `Wind`: 类似于“风”，作为总线和调度器，吹动“灰尘”(`Dust`/任务)运行。
* `Mission`: 任务载体，包含任务名、参数(`Pursuit`)等信息。
* `Worker`: 雪花算法实现，保证高并发下的 ID 唯一性。


* **dust 包**: 示例组件，展示了如何定义一个包含多个方法的结构体(`Dust`)并将其注册到调度器中。
* **brick 包**: 基础数据结构定义。

### 任务控制信号

在任务执行过程中，可以通过 `mission.Name` 传递控制信号：

* `EXIT_FUNCTION`: 退出当前任务。
* `DISCHARGE_CARGO`: 数据卸载/中转（暂未完全开放）。
* `NEW_MISSION`: 开启新任务。
* `INTERRUPT_MISSION`: 中断任务。

## 📄 示例 (Example)

参考 `dust/evan.go` 中的复杂任务交互：

```go
// 模拟一个检查任务
func (d *Dust) CheckIsBig(mission chan *anything.Mission, a []any) {
	for {
        // 在任务内部调用其他任务 "CountXY"
		mis := <-anything.DoChanN("CountXY", []any{rand.Intn(20), rand.Intn(20)})
		
        // 根据返回结果决定逻辑走向
        if mis.Pursuit[0].(int) == 25 {
			mission <- &anything.Mission{Name: "CountXY", Pursuit: []any{"SUCC"}}
			break
		}
	}
}

```

---

*Generated for nihility-dust*