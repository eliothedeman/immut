package immut

import "hash/fnv"

const (
	bits  = 5
	width = 1 << bits
	mask  = width - 1
)

// A trieKey stores both the hashed value and the key that created the value
type trieKeyValue struct {
	hashedKey uint32
	rawKey    []byte
	value     interface{}
}

func (t *trieKeyValue) indexAtDepth(depth uint32) uint32 {
	return (t.hashedKey >> depth) & mask
}

func newTieKeyValue(key []byte, value interface{}) *trieKeyValue {
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
	vals     []*trieKeyValue
	parent   *Trie
	children []*Trie
}

// NewTrie creates and returns a new *Trie
func NewTrie(depth uint32, vals []*trieKeyValue) *Trie {
	return &Trie{
		depth:    depth,
		vals:     vals,
		children: make([]*Trie, bits),
	}
}

// test to see if this key already exists at this level of the trie
func (t *Trie) test(k *trieKeyValue) bool {
	return t.children[k.indexAtDepth(t.depth)] != nil
}
