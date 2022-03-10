package wind

import (
	"errors"
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"net/http"
	"reflect"
	"sync"
	"time"
)

// Wind 实现自Home Nothing
type Wind struct {
	D     []any
	M     map[string]reflect.Value
	C     chan *anything.Mission
	A     sync.Map
	IWork *Worker
}

// Schedule 方法调度器
func (w *Wind) Schedule(startName string, inData ...any) int64 {
	key := w.IWork.GetId()
	go func(I int64) {

		w.C <- &anything.Mission{
			Name:    startName,
			I:       key,
			Pursuit: inData,
		}
		for {
			mission := <-w.C
			fmt.Println("当前方法Schedule：32", key, inData[1].(*http.Request).FormValue("team"), mission.Pursuit)
			switch mission.Name {
			case anything.DC:
				//mission.C = 0
				w.A.Store(I, mission.Pursuit)
			case anything.ExitFunction:
				//mission.C = 1
				w.A.Store(I, mission.Pursuit)
				return
			default:
				go func() {
					defer func() {
						recover()
					}()
					w.M[mission.Name].Call([]reflect.Value{reflect.ValueOf(w.C), reflect.ValueOf(mission.Pursuit)})
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
	w.C = make(chan *anything.Mission, 10)
	w.A = sync.Map{}
	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
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

//panic: interface conversion:
//	interface {} is func(*dust.Dust, chan *anything.Mission, []interface {}),
//				not func(chan *anything.Mission, []interface {})

//你需要先去了解下 golang中   & |  <<  >> 几个符号产生的运算意义
const (
	workerBits  uint8 = 10 //10bit工作机器的id，如果你发现1024台机器不够那就调大次值
	numberBits  uint8 = 12 //12bit 工作序号，如果你发现1毫秒并发生成4096个唯一id不够请调大次值
	workerMax   int64 = -1 ^ (-1 << workerBits)
	numberMax   int64 = -1 ^ (-1 << numberBits)
	timeShift   uint8 = workerBits + numberBits
	workerShift uint8 = numberBits
	// 如果在程序跑了一段时间修改了epoch这个值 可能会导致生成相同的ID，
	//这个值请自行设置为你系统准备上线前的精确到毫秒级别的时间戳，因为雪花时间戳保证唯一的部分最多管69年（2的41次方），
	//所以此值设置为你当前时间戳能够保证你的系统是从当前时间开始往后推69年
	startTime int64 = 1525705533000
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
	//以下表达式才是主菜
	//  (now-startTime)<<timeShift   产生了 41 + （10 + 12）的效应但却并不保证唯一
	//  | (w.workerId << workerShift)  保证了与其他机器不重复
	//  | (w.number))  保证了自己这台机不会重复
	ID := int64((now-startTime)<<timeShift | (w.workerId << workerShift) | (w.number))
	return ID
}
