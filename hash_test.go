package immut

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHashMapPut(t *testing.T) {
	Convey("Given a Hashmap, key and value", t, func() {
		h := NewHashMap()
		k := 72323
		v := "hello world"

		Convey("When the value is stored with the given key", func() {
			h = h.Put(k, v)

			Convey("Expect to retrieve the value", func() {
				nv, found := h.Get(k)
				So(found, ShouldNotEqual, nil)
				So(nv, ShouldEqual, v)
			})
		})

	})
}

func TestIToBytes(t *testing.T) {
	// TODO add more tests for every type and some negative tests
	tests := []struct {
		data interface{}
		want []byte
	}{
		{
			1, []byte{Int, 1, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			1.0, []byte{Float, 0, 0, 0, 0, 0, 0, 240, 63},
		},
	}

	for _, test := range tests {
		got := iToBytes(test.data)
		if !bytes.Equal(test.want, got) {
			t.Errorf("Wanted % x got % x", test.want, got)
		}
	}
}

func TestHashAnything(t *testing.T) {
	tests := []interface{}{
		0, "hello", -1, []byte("warewolf"), 3.2441,
	}

	x := map[uint64]bool{}

	for _, i := range tests {
		y := hashAnything(i)
		if _, inMap := x[y]; inMap {
			t.Fail()
		}
		x[y] = true
	}
}

func TestHashMap(t *testing.T) {

	Convey("Test inserting into a hash map and iterating over it", t, func() {
		h := NewHashMap()
		container := map[int]bool{}
		for i := 0; i < 100; i++ {
			h = h.Put(i, i)
			container[i] = true
		}

		h.Each(func(k, v interface{}) {
			So(k, ShouldEqual, v)
			So(container, ShouldContainKey, k)
			delete(container, k.(int))
		})

		So(len(container), ShouldEqual, 0)

	})
}

func BenchmarkHashAnythingStr(b *testing.B) {
	strs := randStrs(10000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashAnything(strs[i%len(strs)])
	}
}

func BenchmarkHashAnythingInt(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashAnything(3483984)
	}
}

type testByter string

func (t testByter) Bytes() []byte {
	return []byte(t)
}

func BenchmarkHashAnythingByter(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	x := testByter("hello world")
	for i := 0; i < b.N; i++ {
		hashAnything(x)
	}
}
