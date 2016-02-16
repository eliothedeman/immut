package immut

import (
	"encoding/binary"
	"hash/fnv"
	"math"
)

const (
	Int = iota
	UInt
	Float
	String
)

// Byteser returns the []bytes representation of the type. Note this does not need to be able to
// be decoded, just needs to be a unique identifier for the value.
type Byteser interface {
	Bytes() []byte
}

func hashAnything(i interface{}) uint64 {
	v := fnv.New64()
	if x, ok := i.(Byteser); ok {
		v.Write(x.Bytes())
	} else {
		v.Write(iToBytes(i))
	}
	return v.Sum64()
}

func iToBytes(i interface{}) []byte {
	var kind uint8

	// handle strings/bytes
	switch i := i.(type) {
	case string:
		x := make([]byte, len(i)+1)
		x[0] = String
		copy(x[1:], i)
		return x
	case []byte:
		x := make([]byte, len(i)+1)
		x[0] = String
		copy(x[1:], i)
		return i
	}

	// handle numbers
	var x uint64
	found := false
	switch i := i.(type) {
	case int8:
		x = uint64(i)
		kind = Int
		found = true
	case int16:
		x = uint64(i)
		found = true
		kind = Int
	case int32:
		x = uint64(i)
		found = true
		kind = Int
	case int64:
		x = uint64(i)
		found = true
		kind = Int
	case int:
		x = uint64(i)
		found = true
		kind = Int
	case uint8:
		x = uint64(i)
		found = true
		kind = UInt
	case uint16:
		x = uint64(i)
		found = true
		kind = UInt
	case uint32:
		x = uint64(i)
		found = true
		kind = UInt
	case uint64:
		x = uint64(i)
		found = true
		kind = UInt
	case uint:
		x = uint64(i)
		found = true
		kind = UInt
	case float32:
		x = uint64(math.Float32bits(i))
		found = true
		kind = Float
	case float64:
		x = math.Float64bits(i)
		found = true
		kind = Float
	}

	if found {
		b := make([]byte, 9)
		b[0] = kind
		binary.LittleEndian.PutUint64(b[1:], x)
		return b
	}

	return nil
}
