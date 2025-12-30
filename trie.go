package immut

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/maphash"
)

const (
	bitsPerLevel = 2
	width        = 1 << bitsPerLevel // 4 children per node
	maxDepth     = 64 / bitsPerLevel
	ones         = ^uint64(0)
	mask         = ones >> (64 - bitsPerLevel)
)

var (
	seed = maphash.MakeSeed()
)

type Key = comparable
type Val = any
type hashedKey = uint64

// leaf stores a key-value pair
type leaf[K Key, V Val] struct {
	key K
	val V
}

// children is a fixed-size array of 4 child nodes (inlined, no heap allocation)
type children[K Key, V Val] [width]node[K, V]

// node uses inlined children array for memory efficiency with 4-way branching.
type node[K Key, V Val] struct {
	leaf     *leaf[K, V]       // Optional value stored at this node
	children *children[K, V]   // Pointer to inlined children array (nil if no children)
}

// isEmpty returns true if this node has no data
func (n node[K, V]) isEmpty() bool {
	return n.leaf == nil && n.children == nil
}

// hash returns the hash of a key using maphash
func hash[K Key](k K) hashedKey {
	return maphash.Comparable(seed, k)
}

// index extracts the child index from a hash at a given depth
func index(h hashedKey, depth uint) uint {
	shift := 64 - bitsPerLevel*(depth+1)
	return uint((h >> shift) & mask)
}

func (n node[K, V]) insert(k K, v V, h hashedKey, depth uint) node[K, V] {
	// Base case: empty node, create a new leaf
	if n.isEmpty() {
		return node[K, V]{
			leaf: &leaf[K, V]{key: k, val: v},
		}
	}

	// Copy node for immutability
	x := node[K, V]{
		leaf: n.leaf,
	}
	if n.children != nil {
		c := *n.children
		x.children = &c
	}

	// If this node has no leaf, store directly here
	if x.leaf == nil {
		x.leaf = &leaf[K, V]{key: k, val: v}
		return x
	}

	// Same key: update the value
	if x.leaf.key == k {
		x.leaf = &leaf[K, V]{key: k, val: v}
		return x
	}

	// Different key: need to push existing leaf down and insert new key
	// Ensure children array exists
	if x.children == nil {
		x.children = &children[K, V]{}
	}

	// Push existing leaf down into children
	existingHash := hash(x.leaf.key)
	existingIdx := index(existingHash, depth)
	x.children[existingIdx] = x.children[existingIdx].insert(x.leaf.key, x.leaf.val, existingHash, depth+1)
	x.leaf = nil

	// Now insert the new key
	idx := index(h, depth)
	x.children[idx] = x.children[idx].insert(k, v, h, depth+1)
	return x
}

// get retrieves a value from the trie by key
func (n node[K, V]) get(k K, h hashedKey, depth uint) (V, bool) {
	var zero V
	if n.isEmpty() {
		return zero, false
	}

	// Check if this node's leaf matches
	if n.leaf != nil && n.leaf.key == k {
		return n.leaf.val, true
	}

	// No children to search
	if n.children == nil {
		return zero, false
	}

	// Recurse into the appropriate child
	idx := index(h, depth)
	return n.children[idx].get(k, h, depth+1)
}

// delete removes a key from the trie, returning the new trie and whether the key was found
func (n node[K, V]) delete(k K, h hashedKey, depth uint) (node[K, V], bool) {
	if n.isEmpty() {
		return node[K, V]{}, false
	}

	// Check if this node's leaf matches
	if n.leaf != nil && n.leaf.key == k {
		// Found the key - remove the leaf
		return node[K, V]{children: n.children}, true
	}

	// No children to search
	if n.children == nil {
		return n, false
	}

	// Recurse into the appropriate child
	idx := index(h, depth)
	newChild, found := n.children[idx].delete(k, h, depth+1)
	if !found {
		return n, false
	}

	// Copy for immutability
	x := node[K, V]{leaf: n.leaf}
	c := *n.children
	x.children = &c
	x.children[idx] = newChild

	return x, true
}

// count returns the number of key-value pairs in the trie
func (n node[K, V]) count() int {
	if n.isEmpty() {
		return 0
	}

	c := 0
	if n.leaf != nil {
		c = 1
	}

	if n.children != nil {
		for i := range n.children {
			c += n.children[i].count()
		}
	}

	return c
}

// forEach calls fn for each key-value pair in the trie
func (n node[K, V]) forEach(fn func(K, V) bool) bool {
	if n.isEmpty() {
		return true
	}

	if n.leaf != nil {
		if !fn(n.leaf.key, n.leaf.val) {
			return false
		}
	}

	if n.children != nil {
		for i := range n.children {
			if !n.children[i].forEach(fn) {
				return false
			}
		}
	}

	return true
}

// Map is an immutable hash map using a hash array mapped trie (HAMT).
// All operations return a new Map, leaving the original unchanged.
type Map[K Key, V Val] struct {
	root node[K, V]
	len  int
}

// Get retrieves a value by key. Returns the value and true if found,
// or the zero value and false if not found.
func (m Map[K, V]) Get(k K) (V, bool) {
	h := hash(k)
	return m.root.get(k, h, 0)
}

// Set returns a new Map with the key-value pair added or updated.
// The original Map is unchanged.
func (m Map[K, V]) Set(k K, v V) Map[K, V] {
	h := hash(k)
	// Check if key already exists to maintain accurate length
	_, exists := m.root.get(k, h, 0)
	newRoot := m.root.insert(k, v, h, 0)
	newLen := m.len
	if !exists {
		newLen++
	}
	return Map[K, V]{root: newRoot, len: newLen}
}

// Delete returns a new Map with the key removed.
// The original Map is unchanged. Returns the same Map if key not found.
func (m Map[K, V]) Delete(k K) Map[K, V] {
	h := hash(k)
	newRoot, found := m.root.delete(k, h, 0)
	if !found {
		return m
	}
	return Map[K, V]{root: newRoot, len: m.len - 1}
}

// Len returns the number of key-value pairs in the Map.
func (m Map[K, V]) Len() int {
	return m.len
}

// ForEach calls fn for each key-value pair in the Map.
// If fn returns false, iteration stops early.
func (m Map[K, V]) ForEach(fn func(K, V) bool) {
	m.root.forEach(fn)
}

// Has returns true if the key exists in the Map.
func (m Map[K, V]) Has(k K) bool {
	_, ok := m.Get(k)
	return ok
}

// Keys returns a slice of all keys in the Map.
func (m Map[K, V]) Keys() []K {
	keys := make([]K, 0, m.len)
	m.ForEach(func(k K, _ V) bool {
		keys = append(keys, k)
		return true
	})
	return keys
}

// Values returns a slice of all values in the Map.
func (m Map[K, V]) Values() []V {
	vals := make([]V, 0, m.len)
	m.ForEach(func(_ K, v V) bool {
		vals = append(vals, v)
		return true
	})
	return vals
}

// ToMap returns a standard Go map with all key-value pairs.
func (m Map[K, V]) ToMap() map[K]V {
	result := make(map[K]V, m.len)
	m.ForEach(func(k K, v V) bool {
		result[k] = v
		return true
	})
	return result
}

// Constructors

// NewMap creates an empty Map.
func NewMap[K Key, V Val]() Map[K, V] {
	return Map[K, V]{}
}

// MapFrom creates a Map from a standard Go map.
// Uses mutable construction internally for efficiency.
func MapFrom[K Key, V Val](m map[K]V) Map[K, V] {
	b := NewBuilder[K, V]()
	for k, v := range m {
		b.Set(k, v)
	}
	return b.Build()
}

// MapFromPairs creates a Map from alternating key-value pairs.
// Panics if an odd number of arguments is provided.
func MapFromPairs[K Key, V Val](pairs ...any) Map[K, V] {
	if len(pairs)%2 != 0 {
		panic("MapFromPairs requires an even number of arguments")
	}
	b := NewBuilder[K, V]()
	for i := 0; i < len(pairs); i += 2 {
		b.Set(pairs[i].(K), pairs[i+1].(V))
	}
	return b.Build()
}

// Builder provides efficient mutable construction of an immutable Map.
// After calling Build(), the Builder should not be reused.
type Builder[K Key, V Val] struct {
	root node[K, V]
	len  int
}

// NewBuilder creates a new Builder for constructing a Map.
func NewBuilder[K Key, V Val]() *Builder[K, V] {
	return &Builder[K, V]{}
}

// Set adds or updates a key-value pair. Mutates the builder in place.
func (b *Builder[K, V]) Set(k K, v V) *Builder[K, V] {
	h := hash(k)
	_, exists := b.root.get(k, h, 0)
	b.root.insertMut(k, v, h, 0)
	if !exists {
		b.len++
	}
	return b
}

// Delete removes a key. Mutates the builder in place.
func (b *Builder[K, V]) Delete(k K) *Builder[K, V] {
	h := hash(k)
	if deleted := b.root.deleteMut(k, h, 0); deleted {
		b.len--
	}
	return b
}

// Len returns the current number of entries.
func (b *Builder[K, V]) Len() int {
	return b.len
}

// Build returns the constructed Map.
// The Builder should not be used after calling Build.
func (b *Builder[K, V]) Build() Map[K, V] {
	return Map[K, V]{root: b.root, len: b.len}
}

// insertMut mutates the node in place (for builder use only)
func (n *node[K, V]) insertMut(k K, v V, h hashedKey, depth uint) {
	// Empty node - just set the leaf
	if n.isEmpty() {
		n.leaf = &leaf[K, V]{key: k, val: v}
		return
	}

	// No leaf at this node - store directly
	if n.leaf == nil {
		n.leaf = &leaf[K, V]{key: k, val: v}
		return
	}

	// Same key - update value
	if n.leaf.key == k {
		n.leaf = &leaf[K, V]{key: k, val: v}
		return
	}

	// Different key - push existing down and insert new
	// Ensure children array exists
	if n.children == nil {
		n.children = &children[K, V]{}
	}

	// Push existing leaf down
	existingHash := hash(n.leaf.key)
	existingIdx := index(existingHash, depth)
	n.children[existingIdx].insertMut(n.leaf.key, n.leaf.val, existingHash, depth+1)
	n.leaf = nil

	// Insert new key
	idx := index(h, depth)
	n.children[idx].insertMut(k, v, h, depth+1)
}

// deleteMut mutates the node in place (for builder use only)
func (n *node[K, V]) deleteMut(k K, h hashedKey, depth uint) bool {
	if n.isEmpty() {
		return false
	}

	// Check if this node's leaf matches
	if n.leaf != nil && n.leaf.key == k {
		n.leaf = nil
		return true
	}

	// No children to search
	if n.children == nil {
		return false
	}

	// Recurse into appropriate child
	idx := index(h, depth)
	return n.children[idx].deleteMut(k, h, depth+1)
}

// Set Operations

// Union returns a new Map containing all key-value pairs from both maps.
// If a key exists in both, the value from other takes precedence.
func (m Map[K, V]) Union(other Map[K, V]) Map[K, V] {
	result := m
	other.ForEach(func(k K, v V) bool {
		result = result.Set(k, v)
		return true
	})
	return result
}

// Intersection returns a new Map containing only keys present in both maps.
// Values are taken from the receiver (m).
func (m Map[K, V]) Intersection(other Map[K, V]) Map[K, V] {
	var result Map[K, V]
	m.ForEach(func(k K, v V) bool {
		if other.Has(k) {
			result = result.Set(k, v)
		}
		return true
	})
	return result
}

// Difference returns a new Map containing keys from m that are not in other.
func (m Map[K, V]) Difference(other Map[K, V]) Map[K, V] {
	result := m
	other.ForEach(func(k K, _ V) bool {
		result = result.Delete(k)
		return true
	})
	return result
}

// SymmetricDifference returns a new Map containing keys that are in either map but not both.
func (m Map[K, V]) SymmetricDifference(other Map[K, V]) Map[K, V] {
	var result Map[K, V]
	// Add keys from m not in other
	m.ForEach(func(k K, v V) bool {
		if !other.Has(k) {
			result = result.Set(k, v)
		}
		return true
	})
	// Add keys from other not in m
	other.ForEach(func(k K, v V) bool {
		if !m.Has(k) {
			result = result.Set(k, v)
		}
		return true
	})
	return result
}

// Merge returns a new Map with all entries from other added/updated.
// This is an alias for Union.
func (m Map[K, V]) Merge(other Map[K, V]) Map[K, V] {
	return m.Union(other)
}

// Filter returns a new Map containing only entries where fn returns true.
func (m Map[K, V]) Filter(fn func(K, V) bool) Map[K, V] {
	var result Map[K, V]
	m.ForEach(func(k K, v V) bool {
		if fn(k, v) {
			result = result.Set(k, v)
		}
		return true
	})
	return result
}

// Equal returns true if both maps have the same keys and values.
// Values are compared using ==.
func (m Map[K, V]) Equal(other Map[K, V]) bool {
	if m.len != other.len {
		return false
	}
	equal := true
	m.ForEach(func(k K, v V) bool {
		otherV, ok := other.Get(k)
		if !ok || any(v) != any(otherV) {
			equal = false
			return false
		}
		return true
	})
	return equal
}

// MarshalJSON implements json.Marshaler for Map.
// Serializes as a JSON object with string keys (keys must be string-convertible).
func (m Map[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.ToMap())
}

// UnmarshalJSON implements json.Unmarshaler for Map.
// Decodes directly into the trie without intermediate map allocation.
func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	// Expect opening brace
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('{') {
		return fmt.Errorf("expected '{', got %v", tok)
	}

	b := NewBuilder[K, V]()

	// Read key-value pairs
	for dec.More() {
		// Read key (must be string for JSON objects)
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		keyStr, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", keyTok)
		}

		// Convert string key to K
		var key K
		if err := json.Unmarshal([]byte(`"`+keyStr+`"`), &key); err != nil {
			return fmt.Errorf("cannot unmarshal key %q: %w", keyStr, err)
		}

		// Read value
		var val V
		if err := dec.Decode(&val); err != nil {
			return fmt.Errorf("cannot decode value for key %q: %w", keyStr, err)
		}

		b.Set(key, val)
	}

	// Expect closing brace
	tok, err = dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('}') {
		return fmt.Errorf("expected '}', got %v", tok)
	}

	*m = b.Build()
	return nil
}
