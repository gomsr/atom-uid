package generators

import (
	"errors"
	"fmt"
	"github.com/micro-services-roadmap/uid-generator-go/config"
	"github.com/micro-services-roadmap/uid-generator-go/generator"
	"github.com/micro-services-roadmap/uid-generator-go/worker"
	"sync"
	"time"
)

type DefaultConfig struct {
	timeBits   int
	workerBits int
	seqBits    int
	workerId   int64
	epochStr   string
}
type OptionFunc func(v *DefaultConfig)

func TimeBits(timeBits int) OptionFunc {
	return func(config *DefaultConfig) {
		config.timeBits = timeBits
	}
}
func WorkerBits(workerBits int) OptionFunc {
	return func(config *DefaultConfig) {
		config.workerBits = workerBits
	}
}
func SeqBits(seqBits int) OptionFunc {
	return func(config *DefaultConfig) {
		config.seqBits = seqBits
	}
}
func WorkerId(workerId int64) OptionFunc {
	return func(config *DefaultConfig) {
		config.workerId = workerId
	}
}
func EpochStr(epochStr string) OptionFunc {
	return func(config *DefaultConfig) {
		config.epochStr = epochStr
	}
}

// DefaultUidGeneratorV2 represents the UID generator
type DefaultUidGeneratorV2 struct {
	*DefaultConfig
	epochSeconds  int64
	BitsAllocator *generator.BitsAllocator
	sequence      int64
	lastSecond    int64
	mu            sync.Mutex
}

func NewWithConfigV2(conf *config.Config) (*DefaultUidGeneratorV2, error) {
	if conf == nil {
		return nil, errors.New("config is nil")
	}

	wid := conf.IdAssigner.Instance().NextWorkerId()
	return NewWithOptions(TimeBits(conf.TimeBits), WorkerBits(conf.WorkerBits),
		SeqBits(conf.SeqBits), WorkerId(wid), EpochStr(conf.EpochStr))
}

func NewV2(workerId ...int64) (*DefaultUidGeneratorV2, error) {
	var wid int64
	if len(workerId) > 0 {
		wid = workerId[0]
	} else {
		wid = worker.CloudflareWorkerId.Instance().NextWorkerId()
	}

	return NewWithOptions(TimeBits(28), WorkerBits(11), SeqBits(24), WorkerId(wid))
}

// NewWithOptions creates a new DefaultUidGenerator instance
func NewWithOptions(ops ...OptionFunc) (*DefaultUidGeneratorV2, error) {
	//if timeBits+workerBits+seqBits+1 != generator.TotalBits {
	//	return nil, errors.NewDefault("the sum of timeBits, workerBits, and seqBits must be 63")
	//}

	dc := &DefaultConfig{
		timeBits:   28,
		workerBits: 11,
		seqBits:    24,
		workerId:   worker.CloudflareWorkerId.Instance().NextWorkerId(),
	}
	for _, opFunc := range ops {
		opFunc(dc)
	}

	allocator := generator.NewBitsAllocator(dc.timeBits, dc.workerBits, dc.seqBits)
	gtor := &DefaultUidGeneratorV2{
		DefaultConfig: dc,
		BitsAllocator: allocator,
	}

	if len(dc.epochStr) == 0 {
		gtor.epochStr = generator.EpochStr
		dt, _ := time.Parse(generator.EpochStrFormat, generator.EpochStr)
		gtor.epochSeconds = dt.Unix()
		return gtor, nil
	}

	if parse, err := time.Parse(generator.EpochStrFormat, dc.epochStr); err != nil {
		gtor.epochStr = generator.EpochStr
		dt, _ := time.Parse(generator.EpochStrFormat, generator.EpochStr)
		gtor.epochSeconds = dt.Unix()
	} else {
		gtor.epochStr = dc.epochStr
		gtor.epochSeconds = parse.Unix()
	}

	return gtor, nil
}

// GetUID generates a unique ID
func (g *DefaultUidGeneratorV2) GetUID() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.nextId()
}

// MustUID generates a unique ID
func (g *DefaultUidGeneratorV2) MustUID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i := 0; i < 10_000; i++ {
		if id, err := g.nextId(); err == nil {
			return id
		}
	}

	panic("UID generation failed")
}

// nextId generates the next UID
func (g *DefaultUidGeneratorV2) nextId() (int64, error) {
	currentSecond, err := g.getCurrentSecond()
	if err != nil {
		return 0, err
	}

	// Handle clock rollback
	if currentSecond < g.lastSecond {
		refusedSeconds := g.lastSecond - currentSecond
		return 0, fmt.Errorf("clock moved backwards. Refusing for %d seconds", refusedSeconds)
	}

	// Increase sequence at the same second
	if currentSecond == g.lastSecond {
		g.sequence = (g.sequence + 1) & g.BitsAllocator.GetMaxSequence()
		// Exceed sequence max, wait for the next second
		if g.sequence == 0 {
			currentSecond = g.getNextSecond(g.lastSecond)
		}
	} else {
		// Reset sequence if it's a new second
		g.sequence = 0
	}

	g.lastSecond = currentSecond

	// Allocate the bits for UID
	return g.BitsAllocator.Allocate(currentSecond-g.epochSeconds, g.workerId, g.sequence), nil
}

// getCurrentSecond gets the current second
func (g *DefaultUidGeneratorV2) getCurrentSecond() (int64, error) {
	currentSecond := time.Now().Unix()
	if currentSecond-g.epochSeconds > g.BitsAllocator.GetMaxDeltaSeconds() {
		return 0, fmt.Errorf("timestamp bits are exhausted. Refusing UID generation")
	}
	return currentSecond, nil
}

// getNextSecond waits for the next second if the current second is exhausted
func (g *DefaultUidGeneratorV2) getNextSecond(lastTimestamp int64) int64 {
	for {
		timestamp := time.Now().Unix()
		if timestamp > lastTimestamp {
			return timestamp
		}
	}
}

// ParseUID parses a UID and returns its components as a string
func (g *DefaultUidGeneratorV2) ParseUID(uid int64) string {
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
