package wind

import (
	"errors"
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"log"
	"reflect"
	"sync"
	"time"
)

// Wind 实现自Home Nothing
type Wind struct {
	D     []any
	M     map[string]reflect.Value
	C     map[int64]chan *anything.Mission
	A     sync.Map
	E     map[int64]chan struct{}
	IWork *Worker
}

// Schedule 方法调度器
func (w *Wind) Schedule(startName string, inData ...any) int64 {
	key := w.IWork.GetId()
	w.E[key] = make(chan struct{}, 10)
	go func(I int64) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", w.C[key])
				w.E[key] <- struct{}{}
			}
			delete(w.C, I)
		}()
		w.C[I] = make(chan *anything.Mission, 10)
		tempChan := make(chan *anything.Mission, 2)
		w.C[I] <- &anything.Mission{
			Name:    startName,
			Pursuit: inData,
			T:       tempChan,
		}
		for {
			mission := <-w.C[key]
			mission.T = tempChan
			switch mission.Name {
			case anything.DC:
				w.A.Store(I, mission.Pursuit)
			case anything.ExitFunction:
				w.A.Store(I, mission.Pursuit)
				w.E[key] <- struct{}{}
				return
			case anything.NM:
				w.Schedule(mission.Name, mission.Pursuit)
			case anything.IM:
				go func() {
					if err := recover(); err != nil {
						fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", w.C[key])
						w.E[key] <- struct{}{}
					}
					w.M[mission.Pursuit[0].(string)].Call([]reflect.Value{reflect.ValueOf(mission.T), reflect.ValueOf(mission.Pursuit[1:])})
				}()
			case anything.RM:
				log.Println("RM MissionName:")
			default:
				go func() {
					if err := recover(); err != nil {
						fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", w.C[key])
						w.E[key] <- struct{}{}
					}
					w.M[mission.Name].Call([]reflect.Value{reflect.ValueOf(w.C[key]), reflect.ValueOf(mission.Pursuit)})
				}()
			}
		}
	}(key)
	return key
}

// Init 初始化Wind tags:"来无影去无踪"
func (w *Wind) Init() {
	node, err := NewWorker(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IWork = node
	w.M = make(map[string]reflect.Value)
	w.C = make(map[int64]chan *anything.Mission)
	w.E = make(map[int64]chan struct{})
	w.A = sync.Map{}
	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				if _, ok := w.M[method.Name]; ok && method.Name != "Empty" {
					log.Println("panic:", method.Name, "已存在 请检查", client)
				}
				w.M[method.Name] = client.MethodByName(method.Name)
			}
		}
	}
}

func (w *Wind) RegisterRouters(values []reflect.Value) {
	for i := range values {
		dus := values[i].Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.M[method.Name] = values[i].MethodByName(method.Name)
			}
		}
	}
}

// Register 注册方法 根据结构体
// 注：若为指针则会注册所有公开方法 非指针只会注册非指针传递方法
func (w *Wind) Register(a ...any) {
	w.D = append(w.D, a...)
}

const (
	workerBits  uint8 = 10 //10bit工作机器的id，如果你发现1024台机器不够那就调大次值
	numberBits  uint8 = 12 //12bit 工作序号，如果你发现1毫秒并发生成4096个唯一id不够请调大次值
	workerMax   int64 = -1 ^ (-1 << workerBits)
	numberMax   int64 = -1 ^ (-1 << numberBits)
	timeShift         = workerBits + numberBits
	workerShift       = numberBits
	startTime   int64 = 1525705533000
)

type Worker struct {
	mu        sync.Mutex
	timestamp int64
	workerId  int64
	number    int64
}

func NewWorker(workerId int64) (*Worker, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("worker ID excess of quantity")
	}
	// 生成一个新节点
	return &Worker{
		timestamp: 0,
		workerId:  workerId,
		number:    0,
	}, nil
}

func (w *Worker) GetId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now().UnixNano() / 1e6
	if w.timestamp == now {
		w.number++
		if w.number > numberMax {
			for now <= w.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		w.number = 0
		w.timestamp = now
	}
	ID := (now-startTime)<<timeShift | (w.workerId << workerShift) | (w.number)
	return ID
}
