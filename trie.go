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

// A trieKey stores both the hashed value and the key that created the value
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

// A Trie is an immutible implementation of of trie.
// Inspired by Rich Hickey's implementation in clojure.
// Read about it at http://hypirion.com/musings/understanding-persistent-vector-pt-2
type Trie struct {
	depth    uint32
	vals     []Entry
	children [width]*Trie
}

// NewTrie creates and returns a new *Trie
func NewTrie(parent *Trie, vals []Entry) *Trie {
	t := Trie{
		vals: vals,
	}

	if parent != nil {
		t.depth = parent.depth + 1
	}

	return &t
}

// String returns the string representation of the trie
func (t *Trie) String() string {
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

// Put inserts a key, val pair into the trie
func (t *Trie) Put(key []byte, val interface{}) *Trie {
	e := newEntry(key, val)
	return t.put(e)
}

func (t *Trie) put(e Entry) *Trie {

	// the path we use to insert the key
	// these nodes will have to be reallocated
	z := *t
	y := &z
	index := e.indexAtDepth(t.depth)

	// if the slot is open at this level, insert the e
	if y.children[index] == nil {
		// log.Println("Inserting new trie", t.depth)
		y.children[index] = NewTrie(y, []Entry{e})
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

// Get a value from the trie if it exists and (nil, false) if it doesn't
func (t *Trie) Get(key []byte) (interface{}, bool) {
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

// test to see if this key already exists at this level of the trie
func (t *Trie) test(k Entry) bool {

	return t.children[k.indexAtDepth(t.depth)] != nil
}
