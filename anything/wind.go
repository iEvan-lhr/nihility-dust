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
	D []interface{}
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

// OnceSchedule 单步任务调度器
func OnceSchedule(Name string, inData []interface{}) {
	go func() {
		if wind != nil && wind.f != nil {
			do := wind.f.DoMaps()
			defer func() {
				do <- struct{}{}
			}()
		}
		if err := recover(); err != nil {
			log.Println(err)
		}
		load, ok := allMission.Load(Name)
		if ok {
			load.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(inData)})
		} else {
			// 简单模式调度
			load, ok = easyModel.Load(Name)
			if ok {
				// 不支持有返回值的单次调度模式
				load.(reflect.Value).Call(GetReflectValues(inData))
			} else {
				panic("Func is not find:" + Name)
			}
		}
	}()
}

// SchedulePipeline  方法调度器
func SchedulePipeline(Name string, mis chan *Mission, inData []interface{}) {
	if wind != nil && wind.f != nil {
		do := wind.f.DoMaps()
		defer func() {
			do <- struct{}{}
		}()
	}
	load, ok := allMission.Load(Name)
	if ok {
		// 流程控制器调度
		load.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(mis), reflect.ValueOf(inData)})
	} else {
		// 简单模式调度
		load, ok = easyModel.Load(Name)
		if ok {
			call := load.(reflect.Value).Call(GetReflectValues(inData))
			var res []interface{}
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
func AddEasyMission(model []interface{}) {
	for i := range model {
		value := reflect.ValueOf(model[i])
		switch value.Kind() {
		case 19:
			//添加单个方法  无需强制要求入参格式
			name := strings.Split(runtime.FuncForPC(value.Pointer()).Name(), ".")
			easyModel.Store(name[len(name)-1], value)
		case 22:
			//添加结构体的所有方法 要求指针  无需强制要求入参格式
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
func (w *Wind) Schedule(startName string, inData []interface{}) int64 {
	//log.Println(startName)
	var do chan struct{}
	if w.f != nil {
		//注册协程到协程数量控制器中
		do = w.f.DoMaps()
		//log.Println("初始化协程",do)
	}
	//单任务流程唯一键ID
	key := GetId()

	w.E[key] = make(chan struct{}, 10)
	var doFunc func(i int64, name string, data []interface{}, doChan chan struct{})
	//根据KEY 初始化协程
	w.C.Store(key, make(chan *Mission, 10))
	//方法执行器  执行多种状态
	doFunc = func(I int64, name string, data []interface{}, doChan chan struct{}) {
		//log.Println("执行任务",doChan,name)
		//defer 释放操作
		defer func() {
			if err := recover(); err != nil {
				v, o := w.C.Load(I)
				if o {
					fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
				}
			}
			//删除 初始化完成的协程
			w.C.Delete(I)
			if w.f != nil && doChan != nil {
				//回收协程控制器
				doChan <- struct{}{}
			}
			//给出返回值  外部可获取执行完成状态 未支持异步操作
			w.E[I] <- struct{}{}
		}()
		//读取协程 I=KEY
		load, ok := w.C.Load(I)
		if ok {
			//初始化首个任务
			load.(chan *Mission) <- &Mission{
				Name:    name,
				Pursuit: data,
			}
		}
		for {
			//读取协程
			mis, on := w.C.Load(I)
			if !on {
				return
			}
			//收取任务
			mission := <-mis.(chan *Mission)
			// 执行任务
			switch mission.Name {
			case DC:
				//任务数据中转  *暂时未开放*
				w.A.Store(I, mission.Pursuit)
			case ExitFunction:
				//退出任务
				if mission.Pursuit != nil {
					w.A.Store(I, mission.Pursuit)
				}
				return
			case IM:
				// *暂时未开放*
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
					var newDo chan struct{}
					if w.f != nil {
						//注册协程到协程数量控制器中
						newDo = w.f.DoMaps()
					}
					if err := recover(); err != nil {
						v, o := w.C.Load(I)
						if o {
							fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
						}
						w.E[I] <- struct{}{}
						delete(w.E, I)
					}
					//在任务结束后退出KEY
					if mission.Name[0] == '-' {
						mission.Name = mission.Name[1:]
						defer func() {
							mis.(chan *Mission) <- &Mission{Name: ExitFunction, Pursuit: nil}
							newDo <- struct{}{}
						}()
					} else {
						defer func() {
							newDo <- struct{}{}
						}()
					}
					//反射执行任务
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

// Init 初始化Wind 注册所有方法 tags:"来无影去无踪"
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

// RegisterRouters 注册路由
func (w *Wind) RegisterRouters(values []interface{}) {
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
func (w *Wind) Register(a ...interface{}) {
	w.D = append(w.D, a...)
}

// SetController 添加协程数量控制器
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
