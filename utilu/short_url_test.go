package utilu

import (
	"crypto"
	"fmt"
	"testing"
)

func TestCustomUID(t *testing.T) {
	slice := GetUID("https://baidu.com", 10, crypto.SHA1)
	fmt.Println(slice)
}

func BenchmarkMd5UID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Md5UID("https://baidu.com", 7)
	}
}
