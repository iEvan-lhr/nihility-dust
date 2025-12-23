package anything

import (
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// Register 注册方法 根据结构体
func (w *Wind) Register(a ...any) {
	w.D = append(w.D, a...)
}

// RegisterRouters 注册路由
func (w *Wind) RegisterRouters(values []any) {
	w.R = sync.Map{}
	for i := range values {
		client := reflect.ValueOf(values[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.R.Store(method.Name, client.MethodByName(method.Name))
			}
		}
	}
}

// AddEasyMission 添加easyModel中的任务
func AddEasyMission(model []any) {
	for i := range model {
		value := reflect.ValueOf(model[i])
		switch value.Kind() {
		case 19:
			// 添加单个方法
			name := strings.Split(runtime.FuncForPC(value.Pointer()).Name(), ".")
			easyModel.Store(name[len(name)-1], value)
		case 22:
			// 添加结构体的所有方法
			dus := value.Type()
			for j := 0; j < dus.NumMethod(); j++ {
				method := dus.Method(j)
				if method.Name != "" && method.Name != " " {
					if _, ok := easyModel.Load(method.Name); ok && method.Name != "Empty" {
						log.Println("panic:", method.Name, "已存在 请检查", value)
					} else {
						easyModel.Store(method.Name, value.MethodByName(method.Name))
					}
				}
			}
		}
	}
}
