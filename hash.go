package immut

import (
	"encoding/binary"
	"fmt"
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

// HashMap maps anything to anything using the immutible trie type
type HashMap struct {
	keys *Trie
	vals *Trie
}

// NewHashMap
func NewHashMap() *HashMap {
	return &HashMap{
		keys: NewTrie(),
		vals: NewTrie(),
	}
}

// Each funs a function on each k,v pair
func (h *HashMap) Each(f func(k, v interface{})) {
	keys := h.keys.Values()

	for _, k := range keys {
		v, _ := h.Get(k)
		f(k, v)
	}
}

// Keys returns the keys stored in the hash map
func (h *HashMap) Keys() []interface{} {
	return h.keys.Values()
}

// Values returns the values stored in the has map
func (h *HashMap) Values() []interface{} {
	return h.vals.Values()
}

// Put will map anything to anything in the internal trie
func (h *HashMap) Put(k, v interface{}) *HashMap {
	kBytes := iToBytes(k)
	nKeys := h.keys.Put(kBytes, k)
	nVals := h.vals.Put(kBytes, v)

	return &HashMap{
		keys: nKeys,
		vals: nVals,
	}
}

// Get returns the value stored at the given key if it exists else nil, false
func (h *HashMap) Get(k interface{}) (interface{}, bool) {
	return h.vals.Get(iToBytes(k))
}

// Del deletes the value stored at the given key
func (h *HashMap) Del(k interface{}) (*HashMap, interface{}) {
	kBytes := iToBytes(k)
	nKeys, _ := h.keys.Del(kBytes)
	nVals, val := h.vals.Del(kBytes)

	return &HashMap{
		keys: nKeys,
		vals: nVals,
	}, val

}

// IntHashMap maps an int to anything using an immutable trie
type IntHashMap struct {
	t *Trie
}

// Put a kv pair into the map
func (i *IntHashMap) Put(k int64, v interface{}) *IntHashMap {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(k))

	return &IntHashMap{
		t: i.t.Put(b, v),
	}
}

// Get the value stored at the given key
func (i *IntHashMap) Get(k int) (interface{}, bool) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(k))

	return i.t.Get(b)
}

// UintHashMap maps an int to anything using an immutable trie
type UintHashMap struct {
	t *Trie
}

func NewUintHashMap() *UintHashMap {
	return &UintHashMap{
		t: NewTrie(),
	}
}

// Put a kv pair into the map
func (i *UintHashMap) Put(k uint64, v interface{}) *UintHashMap {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, k)

	return &UintHashMap{
		t: i.t.Put(b, v),
	}
}

// Get the value stored at the given key
func (i *UintHashMap) Get(k uint64) (interface{}, bool) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, k)

	return i.t.Get(b)
}

// hashAnything turns anything into a uint64 via a fnv hash
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

	// last resort
	return []byte(fmt.Sprint(i))
}
