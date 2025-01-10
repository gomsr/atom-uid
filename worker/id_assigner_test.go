package worker

import (
	"fmt"
	"testing"
)

func TestType_Instance(t *testing.T) {
	fmt.Println(LocalWorkerId.Instance().NextWorkerId())
}
