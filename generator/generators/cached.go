package generators

import (
	"fmt"
	"github.com/micro-services-roadmap/uid-generator-go/generator"
	"github.com/micro-services-roadmap/uid-generator-go/generator/generators/buffer"
	"log"
	"sync"
	"time"
)

type CachedUidGenerator struct {
	*DefaultUidGenerator
	boostPower            int
	paddingFactor         int
	scheduleInterval      time.Duration
	ringBuffer            *buffer.RingBuffer
	bufferPaddingExecutor *buffer.PaddingExecutor
	rejectedPutHandler    buffer.RejectedPutHandler
	rejectedTakeHandler   buffer.RejectedTakeHandler
	bitsAllocator         *generator.BitsAllocator
	epochSeconds          int64
	workerID              int64
	mu                    sync.Mutex
}

func NewCachedUidGenerator(bitsAllocator *generator.BitsAllocator, workerID int64, epochSeconds int64) *CachedUidGenerator {
	return &CachedUidGenerator{
		boostPower:       3, // Default boost power
		paddingFactor:    buffer.DefaultPaddingPercent,
		bitsAllocator:    bitsAllocator,
		workerID:         workerID,
		epochSeconds:     epochSeconds,
		scheduleInterval: 60 * time.Second, // Default interval
	}
}

//func (g *CachedUidGenerator) AfterPropertiesSet() {
//	g.initRingBuffer()
//	log.Println("Initialized RingBuffer successfully.")
//}

func (g *CachedUidGenerator) GetUID() int64 {
	return g.ringBuffer.Take()
}

func (g *CachedUidGenerator) ParseUID(uid int64) string {
	totalBits := generator.TotalBits
	signBits := g.BitsAllocator.GetSignBits()
	timestampBits := g.BitsAllocator.GetTimestampBits()
	workerIdBits := g.BitsAllocator.GetWorkerIdBits()
	sequenceBits := g.BitsAllocator.GetSequenceBits()

	// Parse UID
	sequence := (uid << uint(totalBits-sequenceBits)) >> uint(totalBits-sequenceBits)
	workerId := (uid << uint(timestampBits+signBits)) >> uint(totalBits-workerIdBits)
	deltaSeconds := uid >> uint(workerIdBits+sequenceBits)

	// Format time from epoch
	thatTime := time.Unix(g.epochSeconds+deltaSeconds, 0)
	return fmt.Sprintf("{\"UID\":\"%d\",\"timestamp\":\"%s\",\"workerId\":\"%d\",\"sequence\":\"%d\"}",
		uid, thatTime.Format("2006-01-02 15:04:05"), workerId, sequence)
}

func (g *CachedUidGenerator) Destroy() {
	g.bufferPaddingExecutor.Shutdown()
}

func (g *CachedUidGenerator) initRingBuffer() {
	bufferSize := (g.bitsAllocator.GetMaxSequence() + 1) << g.boostPower
	g.ringBuffer = buffer.NewRingBuffer(bufferSize, g.paddingFactor)

	log.Printf("Initialized ring buffer size: %d, paddingFactor: %d\n", bufferSize, g.paddingFactor)

	usingSchedule := g.scheduleInterval > 0
	g.bufferPaddingExecutor = buffer.NewBufferPaddingExecutor(g.ringBuffer, &buffer.CachedUidProvider{}, usingSchedule, g.scheduleInterval)
	if usingSchedule {
		g.bufferPaddingExecutor.SetScheduleInterval(g.scheduleInterval)
	}

	log.Printf("Initialized BufferPaddingExecutor. Using schedule: %v, interval: %v\n", usingSchedule, g.scheduleInterval)

	g.ringBuffer.SetBufferPaddingExecutor(g.bufferPaddingExecutor)

	if g.rejectedPutHandler != nil {
		g.ringBuffer.SetRejectedPutHandler(g.rejectedPutHandler)
	}
	if g.rejectedTakeHandler != nil {
		g.ringBuffer.SetRejectedTakeHandler(g.rejectedTakeHandler)
	}

	g.bufferPaddingExecutor.PaddingBuffer()
	g.bufferPaddingExecutor.StartSchedule()
}

func (g *CachedUidGenerator) SetBoostPower(boostPower int) {
	if boostPower <= 0 {
		log.Panic("Boost power must be positive!")
	}
	g.boostPower = boostPower
}

func (g *CachedUidGenerator) SetRejectedPutBufferHandler(handler buffer.RejectedPutHandler) {
	if handler == nil {
		log.Panic("RejectedPutBufferHandler can't be nil!")
	}
	g.rejectedPutHandler = handler
}

func (g *CachedUidGenerator) SetRejectedTakeBufferHandler(handler buffer.RejectedTakeHandler) {
	if handler == nil {
		log.Panic("RejectedTakeBufferHandler can't be nil!")
	}
	g.rejectedTakeHandler = handler
}

func (g *CachedUidGenerator) SetScheduleInterval(interval time.Duration) {
	if interval <= 0 {
		log.Panic("Schedule interval must be positive!")
	}
	g.scheduleInterval = interval
}
