package generators

import (
	"fmt"
	"github.com/micro-services-roadmap/uid-generator-go/generator"
	"github.com/micro-services-roadmap/uid-generator-go/generator/generators/buffer"
	"sync"
	"time"
)

const (
	BoostPower       = 3
	PaddingFactor    = 50
	ScheduleInterval = 60 * time.Second
)

type CachedUidGenerator struct {
	timeBits      int
	workerBits    int
	seqBits       int
	epochStr      string
	epochSeconds  int64
	bitsAllocator *generator.BitsAllocator
	workerId      int64
	sequence      int64
	lastSecond    int64
	mu            sync.Mutex

	boostPower    int
	paddingFactor int
	ringBuffer    *buffer.RingBuffer
}

func NewCached(workerId int64) *CachedUidGenerator {
	return NewCachedUidGenerator(28, 15, 20,
		BoostPower, PaddingFactor, ScheduleInterval, workerId)
}

func NewCachedUidGenerator(timeBits, workerBits, seqBits, boostPower, paddingFactor int,
	scheduleInterval time.Duration, workerId int64, epochStr ...string) *CachedUidGenerator {
	gtor := &CachedUidGenerator{
		timeBits:      timeBits,
		workerBits:    workerBits,
		seqBits:       seqBits,
		bitsAllocator: generator.NewBitsAllocator(timeBits, workerBits, seqBits),
		workerId:      workerId,
		boostPower:    boostPower,
		paddingFactor: paddingFactor,
	}
	// 2. 处理 epochStr
	if len(epochStr) == 0 {
		gtor.epochStr = generator.EpochStr
		dt, _ := time.Parse(generator.EpochStrFormat, generator.EpochStr)
		gtor.epochSeconds = dt.Unix()
	} else {
		if parse, err := time.Parse(generator.EpochStrFormat, epochStr[0]); err != nil {
			gtor.epochStr = generator.EpochStr
			dt, _ := time.Parse(generator.EpochStrFormat, generator.EpochStr)
			gtor.epochSeconds = dt.Unix()
		} else {
			gtor.epochStr = epochStr[0]
			gtor.epochSeconds = parse.Unix()
		}
	}

	// 3. 创建 ringBuffer & 设置拒绝策略 & executor
	bufferSize := int(gtor.bitsAllocator.GetMaxSequence()+1) << boostPower
	ringBuffer := buffer.NewBuffer(int(bufferSize), gtor.paddingFactor)
	fmt.Printf("Initialized ring buffer size: %d, paddingFactor: %d\n", bufferSize, gtor.paddingFactor)

	// 4. 创建 PaddingExecutor
	paddingExecutor := buffer.NewBufferPaddingExecutor(ringBuffer,
		buffer.NewCachedUidProvider(gtor.bitsAllocator), gtor.epochSeconds, scheduleInterval)
	ringBuffer.SetBufferPaddingExecutor(paddingExecutor)
	fmt.Printf("Initialized BufferPaddingExecutor. Using schedule: %v, interval: %v\n", scheduleInterval > 0, scheduleInterval)

	gtor.ringBuffer = ringBuffer
	return gtor
}

func (g *CachedUidGenerator) GetUID() (int64, error) {
	return g.ringBuffer.Take()
}

func (g *CachedUidGenerator) MustUID() int64 {
	take, err := g.ringBuffer.Take()
	if err != nil {
		panic(err)
	}

	return take
}

func (g *CachedUidGenerator) ParseUID(uid int64) string {
	totalBits := generator.TotalBits
	signBits := g.bitsAllocator.GetSignBits()
	timestampBits := g.bitsAllocator.GetTimestampBits()
	workerIdBits := g.bitsAllocator.GetWorkerIdBits()
	sequenceBits := g.bitsAllocator.GetSequenceBits()

	// Parse UID
	sequence := (uid << uint(totalBits-sequenceBits)) >> uint(totalBits-sequenceBits)
	workerId := (uid << uint(timestampBits+signBits)) >> uint(totalBits-workerIdBits)
	deltaSeconds := uid >> uint(workerIdBits+sequenceBits)

	// Format time from epoch
	thatTime := time.Unix(g.epochSeconds+deltaSeconds, 0)
	return fmt.Sprintf("{\"UID\":\"%d\",\"timestamp\":\"%s\",\"workerId\":\"%d\",\"sequence\":\"%d\"}",
		uid, thatTime.Format("2006-01-02 15:04:05"), workerId, sequence)
}

func (g *CachedUidGenerator) SetBoostPower(boostPower int) {
	if boostPower <= 0 {
		fmt.Println("Boost power must be positive!")
	}
	g.boostPower = boostPower
}

func (g *CachedUidGenerator) SetRejectedPutBufferHandler(handler buffer.RejectedPutHandler) {
	if handler == nil {
		fmt.Println("RejectedPutBufferHandler can't be nil!")
	}
	g.ringBuffer.SetRejectedPutHandler(handler)
}

func (g *CachedUidGenerator) SetRejectedTakeBufferHandler(handler buffer.RejectedTakeHandler) {
	if handler == nil {
		fmt.Println("RejectedTakeBufferHandler can't be nil!")
	}
	g.ringBuffer.SetRejectedTakeHandler(handler)
}
