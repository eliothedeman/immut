package immut

import (
	"bytes"
	"fmt"
	"hash/fnv"
)

const (
	bits     = 4
	width    = 1 << bits
	mask     = width - 1
	maxDepth = 16 / bits
)

// A Trie is an immutible implementation of of trie
// Inspired by Rich Hickey's implementation in clojure.
// Read about it at http://hypirion.com/musings/understanding-persistent-vector-pt-2
type Trie struct {
	root *TNode
	size int
}

// Size returns the number of keys/vals in the trie
func (t *Trie) Size() int {
	return t.size
}

// NewTrie creates an empty Trie and returns it
func NewTrie() *Trie {
	return &Trie{
		root: NewTNode(nil, nil),
	}
}

// Put inserts the given value at the given key
func (t *Trie) Put(key []byte, val interface{}) *Trie {
	return &Trie{
		root: t.root.Put(key, val),
		size: t.size + 1,
	}
}

// Get returns the value stored at the given key
func (t *Trie) Get(key []byte) (interface{}, bool) {
	return t.root.Get(key)
}

// Del remove the value stored at the given key and return the value that was stored there
func (t *Trie) Del(key []byte) (*Trie, interface{}) {
	n, i, b := t.root.Del(key)
	if !b {
		return t, nil
	}

	return &Trie{
		root: n,
		size: t.size - 1,
	}, i
}

// Each runs the given function on every k,v pair
func (t *Trie) Each(f func([]byte, interface{})) {
	t.root.Each(f)
}

// Keys returns all of the keys stored in the trie
func (t *Trie) Keys() [][]byte {
	keys := make([][]byte, t.size)
	if t.size == 0 {
		return keys
	}
	count := 0
	t.Each(func(k []byte, v interface{}) {
		keys[count] = k
		count += 1
	})

	return keys
}

// Values returns all fo the values stored in the trie
func (t *Trie) Values() []interface{} {
	values := make([]interface{}, t.size)
	count := 0
	t.Each(func(k []byte, v interface{}) {
		values[count] = v
		count += 1
	})

	return values
}

// A TNodeKey stores both the hashed value and the key that created the value
type Entry struct {
	hashedKey uint32
	rawKey    []byte
	value     interface{}
}

func printBits(u uint32) {
	fmt.Printf("%0b \n", u)
}

func (t Entry) indexAtDepth(depth uint32) uint32 {
	return (t.hashedKey >> (depth)) & mask
}

func (t Entry) sameKey(check Entry) bool {
	return bytes.Equal(t.rawKey, check.rawKey)
}

func newEntry(key []byte, value interface{}) Entry {
	return Entry{
		rawKey:    key,
		hashedKey: hashKey(key),
		value:     value,
	}
}

// hashKey hashes a key into a uint32
func hashKey(key []byte) uint32 {

	// need to convert to a []byte
	h := fnv.New32()
	h.Write(key)
	return h.Sum32()
}

type TNode struct {
	depth    uint32
	vals     []Entry
	children [width]*TNode
}

// NewTNode creates and returns a new *TNode
func NewTNode(parent *TNode, vals []Entry) *TNode {
	t := TNode{
		vals: vals,
	}

	if parent != nil {
		t.depth = parent.depth + 1
	}

	return &t
}

// Each runs a function over all k,v pairs in the node and it's children
func (t *TNode) Each(f func([]byte, interface{})) {
	for _, e := range t.vals {
		f(e.rawKey, e.value)
	}

	// now all children
	for i := 0; i < len(t.children); i++ {
		x := t.children[i]
		if x != nil {
			x.Each(f)
		}
	}
}

// String returns the string representation of the TNode
func (t *TNode) String() string {
	if t == nil {
		return ""
	}
	b := bytes.NewBuffer(nil)
	b.WriteString("{\n")
	for i := 0; i < len(t.children); i++ {
		if t.children[i] != nil {
			b.WriteString(fmt.Sprintf("\t%s\n", t.children[i]))
		}
	}
	b.WriteString(fmt.Sprintf("%v", t.vals))
	b.WriteString("\n}")
	return b.String()
}

// Del remove the value stored at the given key, returns the value if it existed
func (t *TNode) Del(key []byte) (*TNode, interface{}, bool) {
	e := newEntry(key, nil)
	return t.del(e)
}

func (t *TNode) del(e Entry) (*TNode, interface{}, bool) {
	// find the index at the given depth, if it doesn't exist, try to go lower until it does
	z := *t
	y := &z

	// hunt for the key at the current level's values
	for i := 0; i < len(z.vals); i++ {
		if z.vals[i].sameKey(e) {

			// delete in slice
			y.vals = append(y.vals[:i], y.vals[i:]...)
			return y, z.vals[i].value, true
		}
	}
	index := e.indexAtDepth(t.depth)

	if y.children[index] != nil {
		n, i, b := y.children[index].del(e)
		if b {
			y.children[index] = n
			return y, i, b
		}

	}
	// nothing was found
	return t, nil, false
}

// Put inserts a key, val pair into the TNode
func (t *TNode) Put(key []byte, val interface{}) *TNode {
	e := newEntry(key, val)
	return t.put(e)
}

func (t *TNode) put(e Entry) *TNode {

	// the path we use to insert the key
	// these nodes will have to be reallocated
	z := *t
	y := &z
	index := e.indexAtDepth(t.depth)

	// if the slot is open at this level, insert the e
	if y.children[index] == nil {
		// log.Println("Inserting new TNode", t.depth)
		y.children[index] = NewTNode(y, []Entry{e})
		return y
	}

	// if we are at the max depth, start appending
	if y.depth >= maxDepth {
		// log.Println("Appending at ", t.depth)
		y.vals = append(y.vals, e)
		return y
	}

	x := y.children[index]

	// check for a hash collision or that the key already exists
	for i := 0; i < len(x.vals); i++ {
		if x.vals[i].sameKey(e) {
			y.children[index].vals[i] = e
			return y
		}
	}

	y.children[index] = y.children[index].put(e)
	return y
}

// Get a value from the TNode if it exists and (nil, false) if it doesn't
func (t *TNode) Get(key []byte) (interface{}, bool) {
	y := t
	e := newEntry(key, nil)

	// if this part of the hash exists here, go deeper
	y = y.children[e.indexAtDepth(y.depth)]
	for y != nil {

		// go through the list of elements to check to see if it is in here
		for _, v := range y.vals {
			if v.sameKey(e) {
				return v.value, true
			}
		}

		y = y.children[e.indexAtDepth(y.depth)]

	}

	// nothing was found
	return nil, false
}

// test to see if this key already exists at this level of the TNode
func (t *TNode) test(k Entry) bool {

	return t.children[k.indexAtDepth(t.depth)] != nil
}
