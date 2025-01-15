package shorturl

import (
	"testing"
)

var v, _ = New()
var v6, _ = NewV6()
var v7, _ = NewV7()
var v8, _ = NewV8()

func TestDefaultShortUrl_ShortUrl(t *testing.T) {
	uid := v6.MustUID()
	println(uid)

	print(v6.ParseUID(uid))
	println(v6.ShortUrl())
}

func BenchmarkUrl(b *testing.B) { // 59U0BXO1lO
	for i := 0; i < b.N; i++ {
		println(v.ShortUrl())
	}
}

func BenchmarkUrlV6(b *testing.B) { // 2gShXC
	for i := 0; i < b.N; i++ {
		println(v6.ShortUrl())
	}
}

func BenchmarkUrlV7(b *testing.B) { // Am0G1X
	for i := 0; i < b.N; i++ {
		println(v7.ShortUrl())
	}
}

func BenchmarkUrlV8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		println(v8.ShortUrl())
	}
}
