package anything

import (
	"fmt"
	"runtime"
	"strings"

	"log"
	"reflect"
	"sync"
)

// Wind 实现自Home Nothing
type Wind struct {
	D []any
	// M wind 中的func集合
	M sync.Map
	// R wind 中的Router集合
	R sync.Map
	// C wind 中的Channel集合
	C sync.Map
	// A wind 中的返回值集合 在捕获到A中的返回值后 任务流程会默认认为已完成任务流
	A sync.Map
	// E wind 中的外部判段Map 注：不推荐使用异步操作此Map 可能会出现操作异常
	E map[int64]chan struct{}
	// f wind 中的FOX控制器 基于系统阈值来调度任务,均衡调度系统使用,避免出现任务长等待情况,基准测试效率偏差不超过15%
	// 通过 SetController() 方法来设置控制器 需要实现FOX接口
	f FOX
	// IWork 基于雪花闪电算法的唯一键ID生成器 避免任务冲突 无需手动初始化
	// 如需初始化可通过 Wind.IWork=&struct 来设置 必须实现 NewWorker(workerId int64) (*Worker, error) 和 GetId() int64 方法
	IWork *Worker
}

// wind  指针类型的wind 外部不可操作 用于easyModel的操作
var wind *Wind

// allMission wind中的任务map的Copy用于执行异步中的短同步任务
var allMission sync.Map

// easyModel 非路由模式下的注册方法,使用非路由的模式进行执行 在需要高效率的开发过程中不推荐使用这种模式来开发 对代码耦合度较高的环境可以快速解耦
// 可通过不同包中的 init() 方法初始化需要解耦的方法 通过反射模式执行
var easyModel sync.Map

func init() {
	easyModel = sync.Map{}
}

// SchedulePipeline  方法调度器
func SchedulePipeline(Name string, mis chan *Mission, inData []any) {
	if wind != nil && wind.f != nil {
		do := wind.f.DoMaps()
		defer func() {
			do <- struct{}{}
		}()
	}
	load, ok := allMission.Load(Name)
	if ok {
		load.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(mis), reflect.ValueOf(inData)})
	} else {
		load, ok = easyModel.Load(Name)
		if ok {
			call := load.(reflect.Value).Call(GetReflectValues(inData))
			var res []any
			for _, value := range call {
				res = append(res, value.Interface())
			}
			mis <- &Mission{Name: RM, Pursuit: res}
		} else {
			panic("Func is not find:" + Name)
		}
	}
}

// AddEasyMission 添加easyModel中的任务
func AddEasyMission(model []any) {
	for i := range model {
		value := reflect.ValueOf(model[i])
		switch value.Kind() {
		case 19:
			name := strings.Split(runtime.FuncForPC(value.Pointer()).Name(), ".")
			easyModel.Store(name[len(name)-1], value)
		case 22:
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

// Schedule 方法调度器
func (w *Wind) Schedule(startName string, inData []any) int64 {
	var do chan struct{}
	if w.f != nil {
		do = w.f.DoMaps()
	}
	key := GetId()
	w.E[key] = make(chan struct{}, 10)
	var doFunc func(i int64, name string, data []any, doChan chan struct{})
	w.C.Store(key, make(chan *Mission, 10))
	doFunc = func(I int64, name string, data []any, doChan chan struct{}) {
		defer func() {
			if err := recover(); err != nil {
				v, o := w.C.Load(I)
				if o {
					fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
				}
				w.E[I] <- struct{}{}
			}
			w.C.Delete(I)
			if w.f != nil && doChan != nil {
				doChan <- struct{}{}
			}
		}()
		load, ok := w.C.Load(I)
		if ok {
			load.(chan *Mission) <- &Mission{
				Name:    name,
				Pursuit: data,
			}
		}
		for {
			mis, _ := w.C.Load(I)
			mission := <-mis.(chan *Mission)
			switch mission.Name {
			case DC:
				w.A.Store(I, mission.Pursuit)
			case ExitFunction:
				w.A.Store(I, mission.Pursuit)
				w.E[I] <- struct{}{}
				return
			case NM:
				k := GetId()
				mission.T = make(chan *Mission, 2)
				//w.C[k] = mission.T
				w.C.Store(k, mission.T)
				var newDo chan struct{}
				if w.f != nil {
					newDo = w.f.DoMaps()
				}
				go doFunc(k, mission.Pursuit[0].(string), mission.Pursuit[1:], newDo)
			case IM:
				go func() {
					if err := recover(); err != nil {
						v, o := w.C.Load(I)
						if o {
							fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
						}
						w.E[I] <- struct{}{}
					}
					lo, ok1 := w.M.Load(mission.Pursuit[0].(string))
					if ok1 {
						lo.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(mission.T), reflect.ValueOf(mission.Pursuit[1:])})
					}
				}()
			case RM:
				log.Println("RM MissionName:")
			default:
				go func() {
					if err := recover(); err != nil {
						v, o := w.C.Load(I)
						if o {
							fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
						}
						w.E[I] <- struct{}{}
						delete(w.E, I)
					}
					lo, ok1 := w.M.Load(mission.Name)
					if ok1 {
						v, o := w.C.Load(I)
						if o {
							lo.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(v), reflect.ValueOf(mission.Pursuit)})
						}
					}
				}()
			}

		}
	}
	go doFunc(key, startName, inData, do)
	return key
}

// Init 初始化Wind tags:"来无影去无踪"
func (w *Wind) Init() {
	wind = w
	node, err := NewWorker(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IWork = node
	w.M = sync.Map{}
	allMission = sync.Map{}
	w.C = sync.Map{}
	w.E = make(map[int64]chan struct{})
	w.A = sync.Map{}
	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				if _, ok := w.M.Load(method.Name); ok && method.Name != "Empty" {
					log.Println("panic:", method.Name, "已存在 请检查", client)
				}
				w.M.Store(method.Name, client.MethodByName(method.Name))
				allMission.Store(method.Name, client.MethodByName(method.Name))
			}
		}
	}
}

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

// Register 注册方法 根据结构体
// 注：若为指针则会注册所有公开方法 非指针只会注册非指针传递方法
func (w *Wind) Register(a ...any) {
	w.D = append(w.D, a...)
}

func SetController(f FOX) {
	if f != nil {
		if wind == nil {
			wind = &Wind{
				f: f,
			}
		} else {
			wind.f = f
		}
	} else {
		panic("not implemented")
	}
}
