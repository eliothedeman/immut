package immut

import (
	"testing"

	"github.com/eliothedeman/randutil"
	. "github.com/smartystreets/goconvey/convey"
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

func TestTrieDel(t *testing.T) {
	Convey("Ensure deleting keys from one trie doesn't effect the previous generation", t, func() {
		x := NewTrie()

		x = x.Put([]byte("hello"), "world")
		y, i := x.Del([]byte("hello"))
		So(i, ShouldEqual, "world")

		keys := x.Keys()
		for _, k := range keys {
			So("hello", ShouldEqual, string(k))
		}
		keys = y.Keys()
		for _, k := range keys {
			So("hello", ShouldNotEqual, string(k))
		}

	})
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

func BenchmarkTriePutSingle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	x := NewTrie()
	s := []byte("hell world")
	for i := 0; i < b.N; i++ {
		x = x.Put(s, i)
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
