package buffer

import "github.com/micro-services-roadmap/uid-generator-go/generator"

// UidProvider Buffered UID provider(Lambda supported), which provides UID in the same one second
type UidProvider interface {
	// Provide UID in one second
	provide(epochSeconds, momentInSecond int64) []int64
}

func NewCachedUidProvider(g *generator.BitsAllocator) *CachedUidProvider {
	return &CachedUidProvider{g}
}

type CachedUidProvider struct {
	*generator.BitsAllocator
}

// NextIdsForOneSecond Get the UIDs in the same specified second under the max sequence
func (c *CachedUidProvider) provide(epochSeconds, currentSecond int64) []int64 {
	listSize := c.GetMaxSequence() + 1
	uidList := make([]int64, listSize)

	firstSeqUid := c.Allocate(currentSecond-epochSeconds, c.GetMaxWorkerId(), 0)
	for offset := int64(0); offset < listSize; offset++ {
		uidList[offset] = firstSeqUid + offset
	}

	return uidList
}
