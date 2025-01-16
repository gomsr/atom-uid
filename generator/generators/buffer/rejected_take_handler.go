package buffer

import (
	"fmt"
)

// RejectedTakeHandler If cursor catches the tail it means that the ring buffer is empty, any more buffer take request will be rejected.
// Specify the policy to handle the reject. This is a Lambda supported interface
type RejectedTakeHandler interface {
	// Reject take buffer request
	rejectTakeBuffer(ringBuffer *RingBuffer)
}

type PanicTakeBuffer struct{}

func (c *PanicTakeBuffer) rejectTakeBuffer(ringBuffer *RingBuffer) {
	err := fmt.Sprintf("Rejected take buffer. %v", ringBuffer)
	fmt.Println(err)
}
