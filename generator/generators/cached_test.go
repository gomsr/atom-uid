package generators

import (
	"fmt"
	"github.com/gomsr/atom-uid/worker"
	"testing"
	"time"
)

func TestCachedUidGenerator_GetUID(t *testing.T) {
	g := NewCached(worker.LocalWorkerId.Instance().NextWorkerId())
	for i := 0; i < 1000; i++ {
		uid, err := g.GetUID()
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			panic(err)
		}
		fmt.Println(uid)
		fmt.Println(g.ParseUID(uid))
	}
}

func TestParse(t *testing.T) {
	g := NewCached(worker.LocalWorkerId.Instance().NextWorkerId())
	print(g.ParseUID(1132079780664967169))
}
