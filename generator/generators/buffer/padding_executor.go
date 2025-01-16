package buffer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// PaddingExecutor Represents an executor for padding {@link RingBuffer}<br>
//
// There are two kinds of executors: one for scheduled padding, the other for padding immediately.
type PaddingExecutor interface {
	PaddingBuffer()
	AsyncPadding()
	StartSchedule()
	Shutdown()
}

type SchedulePaddingExecutor struct {
	epochSeconds        int64 // 冗余
	running             atomic.Bool
	lastSecond          atomic.Int64
	ringBuffer          *RingBuffer
	uidProvider         UidProvider
	scheduleInterval    time.Duration
	bufferPadSchedule   *time.Ticker
	mu                  sync.WaitGroup
	stopPaddingSchedule chan struct{}
}

func NewBufferPaddingExecutor(ringBuffer *RingBuffer, uidProvider UidProvider, epochSeconds int64, interval time.Duration) *SchedulePaddingExecutor {
	executor := &SchedulePaddingExecutor{
		epochSeconds:        epochSeconds,
		ringBuffer:          ringBuffer,
		uidProvider:         uidProvider,
		scheduleInterval:    interval,
		stopPaddingSchedule: make(chan struct{}),
	}

	executor.lastSecond.Store(time.Now().Unix())
	if interval > 0 {
		executor.bufferPadSchedule = time.NewTicker(interval)
		go executor.StartSchedule()
	}

	executor.PaddingBuffer()

	return executor
}

func (e *SchedulePaddingExecutor) StartSchedule() {
	for {
		select {
		case <-e.bufferPadSchedule.C:
			e.AsyncPadding()
		case <-e.stopPaddingSchedule:
			return
		}
	}
}

func (e *SchedulePaddingExecutor) AsyncPadding() {
	e.mu.Add(1)
	go func() {
		defer e.mu.Done()
		e.PaddingBuffer()
	}()
}

func (e *SchedulePaddingExecutor) PaddingBuffer() {
	fmt.Printf("Ready to padding buffer lastSecond: %d\n", e.lastSecond.Load())

	if !e.running.CompareAndSwap(false, true) {
		fmt.Println("Padding buffer is still running.")
		return
	}
	defer e.running.Store(false)

	isFullRingBuffer := false
	for !isFullRingBuffer {
		uids := e.uidProvider.provide(e.epochSeconds, e.lastSecond.Add(1))
		for _, uid := range uids {
			if !e.ringBuffer.Put(uid) {
				isFullRingBuffer = true
				break
			}
		}
	}

	fmt.Printf("End to padding buffer lastSecond: %d", e.lastSecond.Load())
}

func (e *SchedulePaddingExecutor) Shutdown() {
	if e.bufferPadSchedule != nil {
		close(e.stopPaddingSchedule)
		e.bufferPadSchedule.Stop()
	}
	e.mu.Wait()
}
