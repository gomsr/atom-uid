package buffer

import (
	"errors"
	"fmt"
	"github.com/micro-services-roadmap/uid-generator-go/utilu"
	"sync"
	"sync/atomic"
)

const (
	StartPoint                  = int64(-1)
	CanPutFlag            int32 = 0
	CanTakeFlag           int32 = 1
	DefaultPaddingPercent       = 50
)

// RingBuffer Represents a ring buffer based on array.<br>
// Using array could improve read element performance due to the CUP cache line. To prevent<br/>
// the side effect of False Sharing, {@link PaddedAtomicLong} is using on 'tail' and 'cursor'<p>
//
// A ring buffer is consisted of: <br/>
//   - slots: each element of the array is a slot, which is be set with a UID<br/>
//   - flags: flag array corresponding the same index with the slots, indicates whether can take or put slot<br/>
//   - tail: a sequence of the max slot position to produce<br/>
//   - cursor: a sequence of the min slot position to consume<br/>
type RingBuffer struct {
	bufferSize            int
	indexMask             int64
	slots                 []int64
	flags                 []int32
	tail                  atomic.Int64
	cursor                atomic.Int64
	paddingThreshold      int
	rejectedPutHandler    RejectedPutHandler  // func(rb *RingBuffer, uid int64)
	rejectedTakeHandler   RejectedTakeHandler // func(rb *RingBuffer)
	bufferPaddingExecutor PaddingExecutor     // func()
	mu                    sync.Mutex
}

func NewBuffer(bufferSize int, paddingFactor int) *RingBuffer {
	return NewRingBuffer(bufferSize, paddingFactor, &DiscardPutBuffer{}, &PanicTakeBuffer{}, &SchedulePaddingExecutor{})
}

// NewRingBuffer creates a new RingBuffer
func NewRingBuffer(bufferSize int, paddingFactor int, put RejectedPutHandler, take RejectedTakeHandler, exec *SchedulePaddingExecutor) *RingBuffer {
	if bufferSize <= 0 || bufferSize&(bufferSize-1) != 0 {
		bufferSize = utilu.NextPowerOfTwo(bufferSize)
		fmt.Printf("bufferSize must be a power of 2 and positive, use the next power of two instead: %d\n", bufferSize)
	}
	if paddingFactor <= 0 || paddingFactor >= 100 {
		panic("paddingFactor must be in (0, 100)")
	}

	rb := &RingBuffer{
		bufferSize:            bufferSize,
		indexMask:             int64(bufferSize - 1),
		slots:                 make([]int64, bufferSize),
		flags:                 make([]int32, bufferSize),
		paddingThreshold:      bufferSize * paddingFactor / 100,
		rejectedPutHandler:    put,
		rejectedTakeHandler:   take,
		bufferPaddingExecutor: exec,
	}
	rb.tail.Store(StartPoint)
	rb.cursor.Store(StartPoint)

	// Initialize flags
	for i := range rb.flags {
		rb.flags[i] = CanPutFlag
	}
	return rb
}

// Put an UID in the ring
func (rb *RingBuffer) Put(uid int64) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	currentTail := rb.tail.Load()
	currentCursor := rb.cursor.Load()

	// Check if the buffer is full
	if currentTail-currentCursor == int64(rb.bufferSize)-1 {
		rb.rejectedPutHandler.rejectPutBuffer(rb, uid)
		return false
	}

	// Calculate next slot index
	nextTailIndex := (currentTail) & rb.indexMask
	if !atomic.CompareAndSwapInt32(&rb.flags[nextTailIndex], CanPutFlag, CanTakeFlag) {
		rb.rejectedPutHandler.rejectPutBuffer(rb, uid)
		return false
	}

	// Put UID and update tail
	rb.slots[nextTailIndex] = uid
	rb.tail.Add(1)
	return true
}

// Take an UID from the ring
func (rb *RingBuffer) Take() (int64, error) {
	// 获取当前游标并尝试更新
	currentCursor := rb.cursor.Load()
	nextCursor := rb.cursor.Add(1)
	if nextCursor < currentCursor { // 防止溢出
		return 0, errors.New("cursor can't move back")
	}

	currentTail := rb.tail.Load()
	if nextCursor >= currentTail {
		rb.rejectedTakeHandler.rejectTakeBuffer(rb)
		return 0, errors.New("currentCursor cannot gt currentTail")
	}

	// 异步填充逻辑
	if currentTail-nextCursor < int64(rb.paddingThreshold) {
		fmt.Printf("Reach the padding threshold: %d, tail: %d, cursor: %d, rest: %d", rb.paddingThreshold, currentTail, nextCursor, currentTail-nextCursor)
		rb.bufferPaddingExecutor.AsyncPadding()
	}

	// 获取当前槽位索引并检查状态
	nextCursorIndex := (nextCursor) & rb.indexMask
	uid := rb.slots[nextCursorIndex] // must before swap
	if !atomic.CompareAndSwapInt32(&rb.flags[nextCursorIndex], CanTakeFlag, CanPutFlag) {
		rb.rejectedTakeHandler.rejectTakeBuffer(rb)
		return 0, errors.New("cursor not in can take status")
	}

	// 获取 UID 并更新状态
	return uid, nil
}

// SetRejectedPutHandler sets the handler for rejected put operations
func (rb *RingBuffer) SetRejectedPutHandler(handler RejectedPutHandler) {
	rb.rejectedPutHandler = handler
}

// SetRejectedTakeHandler sets the handler for rejected take operations
func (rb *RingBuffer) SetRejectedTakeHandler(handler RejectedTakeHandler) {
	rb.rejectedTakeHandler = handler
}

// SetBufferPaddingExecutor sets the buffer padding executor
func (rb *RingBuffer) SetBufferPaddingExecutor(executor *SchedulePaddingExecutor) {
	rb.bufferPaddingExecutor = executor
}
