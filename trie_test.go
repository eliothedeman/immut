package immut

import (
	"fmt"
	"testing"

	"github.com/eliothedeman/randutil"
)

func TestTriePutGet(t *testing.T) {
	x := NewTrie(nil, nil)

	x.Put([]byte("hello"), "world")
	fmt.Println(x)
}

func randBytes(count int) [][]byte {
	b := make([][]byte, count)
	for i := 0; i < count; i++ {
		b[i] = []byte(randutil.AlphaString(randutil.IntRange(10, 20)))
	}

	return b
}

func randStrs(count int) []string {
	b := make([]string, count)
	for i := 0; i < count; i++ {
		b[i] = randutil.AlphaString(randutil.IntRange(10, 20))
	}

	return b
}

func BenchmarkTriePut(b *testing.B) {

	strs := randBytes(1000)
	x := NewTrie(nil, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Put(strs[i%len(strs)], randutil.Int())
	}
	fmt.Println(x)
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
