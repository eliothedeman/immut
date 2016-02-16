package immut

import (
	"bytes"
	"testing"
)

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
