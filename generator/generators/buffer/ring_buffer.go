package buffer

import (
	"sync"
	"sync/atomic"
)

const (
	StartPoint                  = -1
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
	indexMask             int
	slots                 []int64
	flags                 []int32
	tail                  int64
	cursor                int64
	paddingThreshold      int
	rejectedPutHandler    RejectedPutHandler  // func(rb *RingBuffer, uid int64)
	rejectedTakeHandler   RejectedTakeHandler // func(rb *RingBuffer)
	bufferPaddingExecutor *PaddingExecutor    // func()
	mu                    sync.Mutex
}

func New(bufferSize int) *RingBuffer {
	return NewRingBuffer(bufferSize, DefaultPaddingPercent)
}

// NewRingBuffer creates a new RingBuffer
func NewRingBuffer(bufferSize int, paddingFactor int) *RingBuffer {
	if bufferSize <= 0 || bufferSize&(bufferSize-1) != 0 {
		panic("bufferSize must be a power of 2 and positive")
	}
	if paddingFactor <= 0 || paddingFactor >= 100 {
		panic("paddingFactor must be in (0, 100)")
	}

	rb := &RingBuffer{
		bufferSize:            bufferSize,
		indexMask:             bufferSize - 1,
		slots:                 make([]int64, bufferSize),
		flags:                 make([]int32, bufferSize),
		tail:                  StartPoint,
		cursor:                StartPoint,
		paddingThreshold:      bufferSize * paddingFactor / 100,
		rejectedPutHandler:    &DiscardPutBuffer{},
		rejectedTakeHandler:   &PanicTakeBuffer{},
		bufferPaddingExecutor: &PaddingExecutor{},
	}

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

	currentTail := atomic.LoadInt64(&rb.tail)
	currentCursor := atomic.LoadInt64(&rb.cursor)

	// Check if the buffer is full
	if currentTail-currentCursor == int64(rb.bufferSize)-1 {
		rb.rejectedPutHandler.rejectPutBuffer(rb, uid)
		return false
	}

	// Calculate next slot index
	nextTailIndex := (currentTail + 1) & int64(rb.indexMask)
	if !atomic.CompareAndSwapInt32(&rb.flags[nextTailIndex], CanPutFlag, CanTakeFlag) {
		rb.rejectedPutHandler.rejectPutBuffer(rb, uid)
		return false
	}

	// Put UID and update tail
	rb.slots[nextTailIndex] = uid
	atomic.AddInt64(&rb.tail, 1)
	return true
}

// Take an UID from the ring
func (rb *RingBuffer) Take() int64 {
	for {
		currentCursor := atomic.LoadInt64(&rb.cursor)
		nextCursor := currentCursor + 1

		if nextCursor > atomic.LoadInt64(&rb.tail) {
			rb.rejectedTakeHandler.rejectTakeBuffer(rb)
		}

		// Calculate next slot index
		nextCursorIndex := int(nextCursor & int64(rb.indexMask))
		if atomic.CompareAndSwapInt32(&rb.flags[nextCursorIndex], CanTakeFlag, CanPutFlag) {
			atomic.StoreInt64(&rb.cursor, nextCursor)
			return rb.slots[nextCursorIndex]
		}
	}
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
func (rb *RingBuffer) SetBufferPaddingExecutor(executor *PaddingExecutor) {
	rb.bufferPaddingExecutor = executor
}
