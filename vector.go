package immut

import "encoding/binary"

// Vector is a constant time lookup with constand time appending and linier prepending
type Vector struct {
	root *TNode
	size int
}

// NewVector returns a new empty vector
func NewVector() *Vector {
	return &Vector{
		root: NewTNode(nil, nil),
	}
}

func newVectorEntry(index int, val interface{}) Entry {
	e := Entry{}
	e.hashedKey = uint32(index)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, e.hashedKey)
	e.rawKey = b
	e.value = val
	return e
}

// Size returns the number of elements in the vector
func (v *Vector) Size() int {
	return v.size
}

// Put the given value at the given index
func (v *Vector) Put(index int, val interface{}) *Vector {
	r := v.root.put(newVectorEntry(index, val))
	return &Vector{
		root: r,
		size: v.size + 1,
	}
}

// Get the value at the givne index
func (v *Vector) Get(index int) (interface{}, bool) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(index))
	return v.root.Get(b)
}

// Slice returns a subslice of the vector.
// same as the built in slice operations mySlice[1:10]
func (v *Vector) Slice(start, end int) *Vector {

	n := NewVector()

	// TODO Don't do all of these allocations
	count := 0
	for i := start; i <= end; i++ {
		x, found := v.Get(i)
		if found {
			n = n.Put(count, x)
		}
		count++
	}
	return n
}
