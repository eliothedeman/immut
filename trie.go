package immut

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"log"
)

const (
	bits  = 4
	width = 1 << bits
	mask  = width - 1
)

// A trieKey stores both the hashed value and the key that created the value
type trieKeyValue struct {
	hashedKey uint32
	rawKey    []byte
	value     interface{}
}

func printBits(u uint32) {
	fmt.Printf("%0b \n", u)
}

func (t *trieKeyValue) indexAtDepth(depth uint32) uint32 {
	return (t.hashedKey >> (depth)) & mask
}

func (t *trieKeyValue) sameKey(check *trieKeyValue) bool {
	return bytes.Equal(t.rawKey, check.rawKey)
}

func newTrieKeyValue(key []byte, value interface{}) *trieKeyValue {
	return &trieKeyValue{
		rawKey:    key,
		hashedKey: hashKey(key),
		value:     value,
	}
}

func hashKey(b []byte) uint32 {
	h := fnv.New32()
	h.Write(b)
	return h.Sum32()
}

// A Trie is an immutible implementation of of trie.
// Inspired by Rich Hickey's implementation in clojure.
// Read about it at http://hypirion.com/musings/understanding-persistent-vector-pt-2
type Trie struct {
	depth    uint32
	vals     *List
	parent   *Trie
	children [width]*Trie
}

// NewTrie creates and returns a new *Trie
func NewTrie(parent *Trie, vals *List) *Trie {
	t := &Trie{
		parent: parent,
		vals:   vals,
	}

	if parent != nil {
		t.depth = parent.depth + 1
	}

	return t
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
	b.WriteString(t.vals.String())
	b.WriteString("\n}")
	return b.String()
}

// IsRoot returns true if this is the root node of the trie
func (t *Trie) IsRoot() bool {
	return t.parent == nil
}

// Put inserts a key, val pair into the trie
func (t *Trie) Put(key []byte, val interface{}) *Trie {

	kv := newTrieKeyValue(key, val)

	// the path we use to insert the key
	// these nodes will have to be reallocated
	y := t
	path := NewList(y)

	// stop at 8 levels deep
	for y.depth < 8 {
		if !y.test(kv) {
			// if the slot is open at this level, insert the kv
			y.children[kv.indexAtDepth(y.depth)] = NewTrie(y, NewList(kv))
			return y
		}
		y = y.children[kv.indexAtDepth(y.depth)]
		path.Prepend(y)
	}

	y.vals.Prepend(kv)

	return y
}

// Get a value from the trie if it exists and (nil, false) if it doesn't
func (t *Trie) Get(key []byte) (interface{}, bool) {
	y := t
	kv := newTrieKeyValue(key, nil)

	// if this part of the hash exists here, go deeper
	y = y.children[kv.indexAtDepth(y.depth)]
	for y != nil {

		// go through the list of elements to check to see if it is in here
		l := y.vals.Filter(func(v *List) bool {
			log.Println(v)
			return v.val.(*trieKeyValue).sameKey(kv)
		})

		if l != nil {
			// if the length of the list is 1, we has found our key
			if l.Len() == 1 {
				return l.val.(*trieKeyValue).value, true
			}
		}

		y = y.children[kv.indexAtDepth(y.depth)]

	}

	// nothing was found
	return nil, false
}

// test to see if this key already exists at this level of the trie
func (t *Trie) test(k *trieKeyValue) bool {

	return t.children[k.indexAtDepth(t.depth)] != nil
}
