package buffer

import (
	"fmt"
)

// RejectedPutHandler If tail catches the cursor it means that the ring buffer is full, any more buffer put request will be rejected.
// Specify the policy to handle the reject. This is a Lambda supported interface
type RejectedPutHandler interface {
	// Reject put buffer request
	rejectPutBuffer(ringBuffer *RingBuffer, uid int64)
}
type DiscardPutBuffer struct{}

func (c *DiscardPutBuffer) rejectPutBuffer(ringBuffer *RingBuffer, uid int64) {
	fmt.Printf("Rejected putting buffer for uid:%d", uid)
}
