package snowflake

import (
	"errors"
	"sync"
	"time"
)

// snowflake 64bit
// 符号位  | 时间戳   | 机器节点ID	| 序列号  |
// 1 bit  | 41 bit  | 10 bit    | 12 bit |

const (
	workerBits  uint8 = 10                      // 机器节点的ID数量，最多 2^10=1024个
	numberBits  uint8 = 12                      // 序列号的ID数量，最多 2^12=4096个
	workerMax   int64 = -1 ^ (-1 << workerBits) // 机器节点的最大值 1023
	numberMax   int64 = -1 ^ (-1 << numberBits) // 序列号的最大值 4095
	timeShift   uint8 = workerBits + numberBits // 时间戳左移偏移量
	workerShift uint8 = numberBits              // 机器节点ID左移偏移量
	startTime         = 1681288475388           // 记录开始时间（毫秒），一旦定义则不能修改，修改后会生成重复ID
)

type Worker struct {
	mu            sync.Mutex // 互斥锁确保并发安全
	lastTimestamp int64      // 记录时间戳，防止时间回拨
	workerId      int64      // 节点ID
	number        int64      // 当前毫秒已生成的序列号，生成累加，一毫秒最多生成2^12=4096个
}

// NewWorker 实例化一个节点
func NewWorker(workerId int64) (*Worker, error) {

	// 检测workerId是否合规
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("workerId ranges from 0 - 1023")
	}

	return &Worker{
		lastTimestamp: 0,
		workerId:      workerId,
		number:        0,
	}, nil
}

// GetId 获取ID
func (w *Worker) GetId() int64 {

	// 加锁、防止并发
	w.mu.Lock()

	// 结束时解锁
	defer w.mu.Unlock()

	// 生成时间戳（毫秒）- 从 go 的 1.17（含）版本后，官方提供了直接获取毫秒,微秒值的方法
	// 1.17 之前版本使用 time.Now().UnixNano() / 1e6
	// now := time.Now().UnixMicro() go这个坑货，返回值是毫秒+纳秒，位数不对
	now := time.Now().UnixNano() / 1e6

	// 判断最后时间是否为当前时间
	var reset bool
	if reset = !(w.lastTimestamp == now); !reset {
		w.number++

		// 判断当前节点1毫秒没生成的ID是否超过了最大序列号限制
		if w.number > numberMax {

			// 如果当前毫秒ID生成已经达到了上限，则等待1毫秒
			for now <= w.lastTimestamp {
				now = time.Now().UnixMicro()
			}
			reset = true
		}
	}

	// 重置数据
	if reset {
		w.number = 0
		w.lastTimestamp = now
	}

	// now - startTime，如果使用当前时间那么时间只能用到2039年，导致浪费，所以做了一个差值处理，比如你的startTime是2023年 那么你就可以用到 2092年
	ID := ((now - startTime) << (timeShift - 0)) | (w.workerId << workerShift) | (w.number)
	return ID
}

func (w *Worker) GetLastTimestamp() int64 {
	return w.lastTimestamp
}

func (w *Worker) GetWorkerId() int64 {
	return w.workerId
}
