package buffer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// PaddingExecutor Represents an executor for padding {@link RingBuffer}<br>
//
//	There are two kinds of executors: one for scheduled padding, the other for padding immediately.
type PaddingExecutor struct {
	running             atomic.Bool
	lastSecond          atomic.Int64
	ringBuffer          *RingBuffer
	uidProvider         UidProvider
	bufferPadExecutors  sync.WaitGroup
	bufferPadSchedule   *time.Ticker
	scheduleInterval    time.Duration
	stopPaddingSchedule chan struct{}
}

func NewBufferPaddingExecutor(ringBuffer *RingBuffer, uidProvider UidProvider, usingSchedule bool, interval time.Duration) *PaddingExecutor {
	executor := &PaddingExecutor{
		ringBuffer:          ringBuffer,
		uidProvider:         uidProvider,
		scheduleInterval:    interval,
		stopPaddingSchedule: make(chan struct{}),
	}

	executor.lastSecond.Store(time.Now().Unix())

	if usingSchedule {
		executor.bufferPadSchedule = time.NewTicker(interval)
		go executor.StartSchedule()
	}

	return executor
}

func (e *PaddingExecutor) StartSchedule() {
	for {
		select {
		case <-e.bufferPadSchedule.C:
			e.AsyncPadding()
		case <-e.stopPaddingSchedule:
			return
		}
	}
}

func (e *PaddingExecutor) AsyncPadding() {
	e.bufferPadExecutors.Add(1)
	go func() {
		defer e.bufferPadExecutors.Done()
		e.PaddingBuffer()
	}()
}

func (e *PaddingExecutor) PaddingBuffer() {
	fmt.Printf("Ready to padding buffer lastSecond: %d", e.lastSecond.Load())

	if !e.running.CompareAndSwap(false, true) {
		fmt.Println("Padding buffer is still running.")
		return
	}
	defer e.running.Store(false)

	isFullRingBuffer := false
	for !isFullRingBuffer {
		uids := e.uidProvider.provide(e.lastSecond.Add(1))
		for _, uid := range uids {
			if !e.ringBuffer.Put(uid) {
				isFullRingBuffer = true
				break
			}
		}
	}

	fmt.Printf("End to padding buffer lastSecond: %d", e.lastSecond.Load())
}

func (e *PaddingExecutor) Shutdown() {
	if e.bufferPadSchedule != nil {
		close(e.stopPaddingSchedule)
		e.bufferPadSchedule.Stop()
	}
	e.bufferPadExecutors.Wait()
}

func (e *PaddingExecutor) SetScheduleInterval(interval time.Duration) {
	if interval <= 0 {
		panic("Schedule interval must be positive!")
	}
	e.scheduleInterval = interval
	if e.bufferPadSchedule != nil {
		e.bufferPadSchedule.Reset(interval)
	}
}
