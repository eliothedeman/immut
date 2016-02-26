package immut

import (
	"testing"

	"github.com/eliothedeman/randutil"
)

func TestTriePutGet(t *testing.T) {
	x := NewTrie()

	y := x.Put([]byte("hello"), "world")
	if _, found := x.Get([]byte("hello")); found {
		t.Error("Persistance broken. Hello should not have been found")
	}

	if out, found := y.Get([]byte("hello")); !found || out.(string) != "world" {
		t.Fail()
	}
}

func randStrs(count int) []string {
	b := make([]string, count)
	for i := 0; i < count; i++ {
		b[i] = randutil.AlphaString(randutil.IntRange(10, 20))
	}

	return b
}

func randBytes(count int) [][]byte {
	b := make([][]byte, count)
	x := randStrs(count)
	for i := 0; i < count; i++ {
		b[i] = []byte(x[i])
	}
	return b
}

func BenchmarkTriePut(b *testing.B) {

	strs := randBytes(1000)
	x := NewTrie()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = x.Put(strs[i%len(strs)], randutil.Int())
	}
}

func BenchmarkHashPut(b *testing.B) {
	strs := randStrs(1000)
	x := make(map[string]int)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x[strs[i%len(strs)]] = randutil.Int()
	}
}
