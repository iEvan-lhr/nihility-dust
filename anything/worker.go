package anything

import (
	"errors"
	"sync"
	"time"
)

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

func GetId() int64 {
	if w == nil {
		w, _ = NewWorker(1)
	}
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

var w *Worker
