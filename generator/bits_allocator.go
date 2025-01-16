package generator

import (
	"fmt"
)

const (
	TotalBits = 1 << 6
)

// BitsAllocator is responsible for allocating the 64 bits for UID
type BitsAllocator struct {
	signBits      int
	timestampBits int
	workerIdBits  int
	sequenceBits  int

	maxDeltaSeconds int64
	maxWorkerId     int64
	maxSequence     int64

	timestampShift int
	workerIdShift  int
}

// NewBitsAllocator creates a new BitsAllocator with the specified bit lengths
func NewBitsAllocator(timestampBits, workerIdBits, sequenceBits int) *BitsAllocator {
	// Ensure we allocate 64 bits
	//totalBits := 1 + timestampBits + workerIdBits + sequenceBits
	//if totalBits != TotalBits {
	//	panic("Total bits do not add up to 64")
	//}

	maxDeltaSeconds := ^(-1 << uint(timestampBits))
	maxWorkerId := ^(-1 << uint(workerIdBits))
	maxSequence := ^(-1 << uint(sequenceBits))

	timestampShift := workerIdBits + sequenceBits
	workerIdShift := sequenceBits

	return &BitsAllocator{
		signBits:        1,
		timestampBits:   timestampBits,
		workerIdBits:    workerIdBits,
		sequenceBits:    sequenceBits,
		maxDeltaSeconds: int64(maxDeltaSeconds),
		maxWorkerId:     int64(maxWorkerId),
		maxSequence:     int64(maxSequence),
		timestampShift:  timestampShift,
		workerIdShift:   workerIdShift,
	}
}

// Allocate combines the delta seconds, worker ID, and sequence into a single UID
func (b *BitsAllocator) Allocate(deltaSeconds, workerId, sequence int64) int64 {
	return (deltaSeconds << uint(b.timestampShift)) | (workerId << uint(b.workerIdShift)) | sequence
}

// Getters for all the fields in BitsAllocator
func (b *BitsAllocator) GetSignBits() int          { return b.signBits }
func (b *BitsAllocator) GetTimestampBits() int     { return b.timestampBits }
func (b *BitsAllocator) GetWorkerIdBits() int      { return b.workerIdBits }
func (b *BitsAllocator) GetSequenceBits() int      { return b.sequenceBits }
func (b *BitsAllocator) GetMaxDeltaSeconds() int64 { return b.maxDeltaSeconds }
func (b *BitsAllocator) GetMaxWorkerId() int64     { return b.maxWorkerId }
func (b *BitsAllocator) GetMaxSequence() int64     { return b.maxSequence }
func (b *BitsAllocator) GetTimestampShift() int    { return b.timestampShift }
func (b *BitsAllocator) GetWorkerIdShift() int     { return b.workerIdShift }

// String provides a string representation of BitsAllocator
func (b *BitsAllocator) String() string {
	return fmt.Sprintf("bitsAllocator{signBits: %d, timestampBits: %d, workerIdBits: %d, sequenceBits: %d, "+
		"maxDeltaSeconds: %d, maxWorkerId: %d, maxSequence: %d, timestampShift: %d, workerIdShift: %d}",
		b.signBits, b.timestampBits, b.workerIdBits, b.sequenceBits, b.maxDeltaSeconds, b.maxWorkerId,
		b.maxSequence, b.timestampShift, b.workerIdShift)
}
