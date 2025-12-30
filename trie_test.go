package immut

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestInsertAndGet(t *testing.T) {
	var root node[string, int]

	// Insert a single key
	h := hash("hello")
	root = root.insert("hello", 42, h, 0)

	// Retrieve it
	val, ok := root.get("hello", h, 0)
	if !ok {
		t.Fatal("expected to find 'hello'")
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}

	// Key not found
	h2 := hash("world")
	_, ok = root.get("world", h2, 0)
	if ok {
		t.Error("expected 'world' to not be found")
	}
}

func TestInsertOverwrite(t *testing.T) {
	var root node[string, int]

	h := hash("key")
	root = root.insert("key", 1, h, 0)
	root = root.insert("key", 2, h, 0)

	val, ok := root.get("key", h, 0)
	if !ok {
		t.Fatal("expected to find 'key'")
	}
	if val != 2 {
		t.Errorf("expected 2 after overwrite, got %d", val)
	}
}

func TestImmutability(t *testing.T) {
	var root node[string, int]

	h1 := hash("a")
	root1 := root.insert("a", 1, h1, 0)

	h2 := hash("b")
	root2 := root1.insert("b", 2, h2, 0)

	// root1 should still only have "a"
	val, ok := root1.get("a", h1, 0)
	if !ok || val != 1 {
		t.Error("root1 should have 'a' = 1")
	}

	_, ok = root1.get("b", h2, 0)
	if ok {
		t.Error("root1 should NOT have 'b' (immutability violated)")
	}

	// root2 should have both
	val, ok = root2.get("a", h1, 0)
	if !ok || val != 1 {
		t.Error("root2 should have 'a' = 1")
	}

	val, ok = root2.get("b", h2, 0)
	if !ok || val != 2 {
		t.Error("root2 should have 'b' = 2")
	}
}

func TestMultipleInserts(t *testing.T) {
	var root node[string, int]

	keys := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape"}

	for i, k := range keys {
		h := hash(k)
		root = root.insert(k, i, h, 0)
	}

	// Verify all keys are retrievable
	for i, k := range keys {
		h := hash(k)
		val, ok := root.get(k, h, 0)
		if !ok {
			t.Errorf("expected to find key %q", k)
			continue
		}
		if val != i {
			t.Errorf("key %q: expected %d, got %d", k, i, val)
		}
	}
}

func TestManyKeys(t *testing.T) {
	var root node[int, int]

	n := 1000
	for i := range n {
		h := hash(i)
		root = root.insert(i, i*10, h, 0)
	}

	// Verify all keys
	for i := range n {
		h := hash(i)
		val, ok := root.get(i, h, 0)
		if !ok {
			t.Errorf("expected to find key %d", i)
			continue
		}
		if val != i*10 {
			t.Errorf("key %d: expected %d, got %d", i, i*10, val)
		}
	}

	// Verify missing keys
	for i := n; i < n+100; i++ {
		h := hash(i)
		_, ok := root.get(i, h, 0)
		if ok {
			t.Errorf("key %d should not exist", i)
		}
	}
}

func TestHashCollisionHandling(t *testing.T) {
	// Insert many keys that may have partial hash collisions
	var root node[string, int]

	for i := range 100 {
		k := fmt.Sprintf("key%d", i)
		h := hash(k)
		root = root.insert(k, i, h, 0)
	}

	// All should be retrievable
	for i := range 100 {
		k := fmt.Sprintf("key%d", i)
		h := hash(k)
		val, ok := root.get(k, h, 0)
		if !ok {
			t.Errorf("expected to find %q", k)
			continue
		}
		if val != i {
			t.Errorf("%q: expected %d, got %d", k, i, val)
		}
	}
}

// Tests for public Map API

func TestMapGetSet(t *testing.T) {
	var m Map[string, int]

	// Empty map
	_, ok := m.Get("foo")
	if ok {
		t.Error("expected empty map to not find key")
	}

	// Set and get
	m = m.Set("foo", 42)
	val, ok := m.Get("foo")
	if !ok || val != 42 {
		t.Errorf("expected foo=42, got %d, %v", val, ok)
	}

	// Update existing key
	m2 := m.Set("foo", 100)
	val, _ = m2.Get("foo")
	if val != 100 {
		t.Errorf("expected foo=100 after update, got %d", val)
	}

	// Original unchanged (immutability)
	val, _ = m.Get("foo")
	if val != 42 {
		t.Errorf("original should still have foo=42, got %d", val)
	}
}

func TestMapLen(t *testing.T) {
	var m Map[string, int]

	if m.Len() != 0 {
		t.Errorf("expected len 0, got %d", m.Len())
	}

	m = m.Set("a", 1)
	if m.Len() != 1 {
		t.Errorf("expected len 1, got %d", m.Len())
	}

	m = m.Set("b", 2)
	if m.Len() != 2 {
		t.Errorf("expected len 2, got %d", m.Len())
	}

	// Update existing key shouldn't change length
	m = m.Set("a", 100)
	if m.Len() != 2 {
		t.Errorf("expected len 2 after update, got %d", m.Len())
	}
}

func TestMapDelete(t *testing.T) {
	var m Map[string, int]
	m = m.Set("a", 1).Set("b", 2).Set("c", 3)

	if m.Len() != 3 {
		t.Errorf("expected len 3, got %d", m.Len())
	}

	// Delete existing key
	m2 := m.Delete("b")
	if m2.Len() != 2 {
		t.Errorf("expected len 2 after delete, got %d", m2.Len())
	}

	_, ok := m2.Get("b")
	if ok {
		t.Error("expected 'b' to be deleted")
	}

	// Original unchanged
	val, ok := m.Get("b")
	if !ok || val != 2 {
		t.Error("original should still have 'b'")
	}

	// Delete non-existent key
	m3 := m2.Delete("nonexistent")
	if m3.Len() != 2 {
		t.Errorf("delete of non-existent key should not change len")
	}
}

func TestMapForEach(t *testing.T) {
	var m Map[string, int]
	m = m.Set("a", 1).Set("b", 2).Set("c", 3)

	seen := make(map[string]int)
	m.ForEach(func(k string, v int) bool {
		seen[k] = v
		return true
	})

	if len(seen) != 3 {
		t.Errorf("expected 3 items, got %d", len(seen))
	}

	for _, k := range []string{"a", "b", "c"} {
		expected, _ := m.Get(k)
		if seen[k] != expected {
			t.Errorf("ForEach: %s expected %d, got %d", k, expected, seen[k])
		}
	}
}

func TestMapForEachEarlyStop(t *testing.T) {
	var m Map[int, int]
	for i := range 100 {
		m = m.Set(i, i)
	}

	count := 0
	m.ForEach(func(k, v int) bool {
		count++
		return count < 10 // Stop after 10
	})

	if count != 10 {
		t.Errorf("expected ForEach to stop after 10, got %d", count)
	}
}

func TestMapManyOperations(t *testing.T) {
	var m Map[int, int]

	// Insert many
	for i := range 1000 {
		m = m.Set(i, i*10)
	}

	if m.Len() != 1000 {
		t.Errorf("expected len 1000, got %d", m.Len())
	}

	// Verify all
	for i := range 1000 {
		val, ok := m.Get(i)
		if !ok || val != i*10 {
			t.Errorf("key %d: expected %d, got %d", i, i*10, val)
		}
	}

	// Delete half
	for i := 0; i < 500; i++ {
		m = m.Delete(i)
	}

	if m.Len() != 500 {
		t.Errorf("expected len 500 after deletes, got %d", m.Len())
	}

	// Verify remaining
	for i := 500; i < 1000; i++ {
		val, ok := m.Get(i)
		if !ok || val != i*10 {
			t.Errorf("key %d should still exist", i)
		}
	}

	// Verify deleted
	for i := range 500 {
		_, ok := m.Get(i)
		if ok {
			t.Errorf("key %d should be deleted", i)
		}
	}
}

func TestMapHas(t *testing.T) {
	m := NewMap[string, int]().Set("a", 1).Set("b", 2)

	if !m.Has("a") {
		t.Error("expected Has('a') to be true")
	}
	if !m.Has("b") {
		t.Error("expected Has('b') to be true")
	}
	if m.Has("c") {
		t.Error("expected Has('c') to be false")
	}
}

func TestMapKeysValues(t *testing.T) {
	m := NewMap[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)

	keys := m.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	vals := m.Values()
	if len(vals) != 3 {
		t.Errorf("expected 3 values, got %d", len(vals))
	}

	// Check all keys exist
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !keySet[expected] {
			t.Errorf("missing key %s", expected)
		}
	}
}

func TestMapFromAndToMap(t *testing.T) {
	stdMap := map[string]int{"a": 1, "b": 2, "c": 3}

	m := MapFrom(stdMap)
	if m.Len() != 3 {
		t.Errorf("expected len 3, got %d", m.Len())
	}

	for k, v := range stdMap {
		got, ok := m.Get(k)
		if !ok || got != v {
			t.Errorf("key %s: expected %d, got %d", k, v, got)
		}
	}

	// Convert back
	result := m.ToMap()
	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
	for k, v := range stdMap {
		if result[k] != v {
			t.Errorf("ToMap: key %s expected %d, got %d", k, v, result[k])
		}
	}
}

func TestMapUnion(t *testing.T) {
	m1 := NewMap[string, int]().Set("a", 1).Set("b", 2)
	m2 := NewMap[string, int]().Set("b", 20).Set("c", 3)

	result := m1.Union(m2)

	if result.Len() != 3 {
		t.Errorf("expected len 3, got %d", result.Len())
	}

	// "a" from m1
	if v, _ := result.Get("a"); v != 1 {
		t.Errorf("expected a=1, got %d", v)
	}
	// "b" from m2 (overrides m1)
	if v, _ := result.Get("b"); v != 20 {
		t.Errorf("expected b=20, got %d", v)
	}
	// "c" from m2
	if v, _ := result.Get("c"); v != 3 {
		t.Errorf("expected c=3, got %d", v)
	}
}

func TestMapIntersection(t *testing.T) {
	m1 := NewMap[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
	m2 := NewMap[string, int]().Set("b", 20).Set("c", 30).Set("d", 4)

	result := m1.Intersection(m2)

	if result.Len() != 2 {
		t.Errorf("expected len 2, got %d", result.Len())
	}

	// Values from m1
	if v, _ := result.Get("b"); v != 2 {
		t.Errorf("expected b=2 (from m1), got %d", v)
	}
	if v, _ := result.Get("c"); v != 3 {
		t.Errorf("expected c=3 (from m1), got %d", v)
	}

	if result.Has("a") || result.Has("d") {
		t.Error("intersection should not have 'a' or 'd'")
	}
}

func TestMapDifference(t *testing.T) {
	m1 := NewMap[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
	m2 := NewMap[string, int]().Set("b", 20).Set("d", 4)

	result := m1.Difference(m2)

	if result.Len() != 2 {
		t.Errorf("expected len 2, got %d", result.Len())
	}

	if !result.Has("a") || !result.Has("c") {
		t.Error("difference should have 'a' and 'c'")
	}
	if result.Has("b") {
		t.Error("difference should not have 'b'")
	}
}

func TestMapSymmetricDifference(t *testing.T) {
	m1 := NewMap[string, int]().Set("a", 1).Set("b", 2)
	m2 := NewMap[string, int]().Set("b", 20).Set("c", 3)

	result := m1.SymmetricDifference(m2)

	if result.Len() != 2 {
		t.Errorf("expected len 2, got %d", result.Len())
	}

	if !result.Has("a") || !result.Has("c") {
		t.Error("symmetric difference should have 'a' and 'c'")
	}
	if result.Has("b") {
		t.Error("symmetric difference should not have 'b'")
	}
}

func TestMapFilter(t *testing.T) {
	m := NewMap[string, int]().Set("a", 1).Set("b", 2).Set("c", 3).Set("d", 4)

	// Keep only even values
	result := m.Filter(func(k string, v int) bool {
		return v%2 == 0
	})

	if result.Len() != 2 {
		t.Errorf("expected len 2, got %d", result.Len())
	}

	if !result.Has("b") || !result.Has("d") {
		t.Error("filter should keep 'b' and 'd'")
	}
	if result.Has("a") || result.Has("c") {
		t.Error("filter should remove 'a' and 'c'")
	}
}

func TestMapEqual(t *testing.T) {
	m1 := NewMap[string, int]().Set("a", 1).Set("b", 2)
	m2 := NewMap[string, int]().Set("a", 1).Set("b", 2)
	m3 := NewMap[string, int]().Set("a", 1).Set("b", 3)
	m4 := NewMap[string, int]().Set("a", 1)

	if !m1.Equal(m2) {
		t.Error("m1 and m2 should be equal")
	}

	if m1.Equal(m3) {
		t.Error("m1 and m3 should not be equal (different value)")
	}

	if m1.Equal(m4) {
		t.Error("m1 and m4 should not be equal (different length)")
	}
}

func TestMapMerge(t *testing.T) {
	m1 := NewMap[string, int]().Set("a", 1)
	m2 := NewMap[string, int]().Set("b", 2)

	result := m1.Merge(m2)
	if result.Len() != 2 {
		t.Errorf("expected len 2, got %d", result.Len())
	}
}

func TestBuilder(t *testing.T) {
	b := NewBuilder[string, int]()
	b.Set("a", 1).Set("b", 2).Set("c", 3)

	if b.Len() != 3 {
		t.Errorf("expected len 3, got %d", b.Len())
	}

	m := b.Build()

	if m.Len() != 3 {
		t.Errorf("expected map len 3, got %d", m.Len())
	}

	for _, k := range []string{"a", "b", "c"} {
		if !m.Has(k) {
			t.Errorf("expected key %s", k)
		}
	}
}

func TestBuilderUpdate(t *testing.T) {
	b := NewBuilder[string, int]()
	b.Set("a", 1).Set("a", 2)

	if b.Len() != 1 {
		t.Errorf("expected len 1 after update, got %d", b.Len())
	}

	m := b.Build()
	v, _ := m.Get("a")
	if v != 2 {
		t.Errorf("expected a=2, got %d", v)
	}
}

func TestBuilderDelete(t *testing.T) {
	b := NewBuilder[string, int]()
	b.Set("a", 1).Set("b", 2).Set("c", 3)
	b.Delete("b")

	if b.Len() != 2 {
		t.Errorf("expected len 2 after delete, got %d", b.Len())
	}

	m := b.Build()
	if m.Has("b") {
		t.Error("expected 'b' to be deleted")
	}
}

func TestBuilderManyKeys(t *testing.T) {
	b := NewBuilder[int, int]()
	n := 1000
	for i := range n {
		b.Set(i, i*10)
	}

	if b.Len() != n {
		t.Errorf("expected len %d, got %d", n, b.Len())
	}

	m := b.Build()
	for i := range n {
		v, ok := m.Get(i)
		if !ok || v != i*10 {
			t.Errorf("key %d: expected %d, got %d", i, i*10, v)
		}
	}
}

func TestMapJSON(t *testing.T) {
	m := NewMap[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)

	// Marshal
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal
	var m2 Map[string, int]
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify
	if m2.Len() != 3 {
		t.Errorf("expected len 3, got %d", m2.Len())
	}

	if !m.Equal(m2) {
		t.Error("unmarshaled map should equal original")
	}
}

func TestMapJSONRoundTrip(t *testing.T) {
	original := MapFrom(map[string]string{
		"name":  "test",
		"value": "hello",
	})

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored Map[string, string]
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !original.Equal(restored) {
		t.Error("round-trip should preserve equality")
	}
}

// Benchmarks comparing immutable trie vs built-in map

var sizes = []int{100, 1000, 10000}

// BenchmarkTrieInsert measures insert performance for the immutable trie
func BenchmarkTrieInsert(b *testing.B) {
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				var root node[int, int]
				for i := range size {
					h := hash(i)
					root = root.insert(i, i, h, 0)
				}
			}
		})
	}
}

// BenchmarkBuilderInsert measures insert performance using the mutable Builder
func BenchmarkBuilderInsert(b *testing.B) {
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				builder := NewBuilder[int, int]()
				for i := range size {
					builder.Set(i, i)
				}
				_ = builder.Build()
			}
		})
	}
}

// BenchmarkMapInsert measures insert performance for built-in map
func BenchmarkMapInsert(b *testing.B) {
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				m := make(map[int]int)
				for i := range size {
					m[i] = i
				}
			}
		})
	}
}

// BenchmarkTrieGet measures lookup performance for the immutable trie
func BenchmarkTrieGet(b *testing.B) {
	for _, size := range sizes {
		// Pre-build the trie
		var root node[int, int]
		for i := range size {
			h := hash(i)
			root = root.insert(i, i, h, 0)
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				for i := range size {
					h := hash(i)
					root.get(i, h, 0)
				}
			}
		})
	}
}

// BenchmarkMapGet measures lookup performance for built-in map
func BenchmarkMapGet(b *testing.B) {
	for _, size := range sizes {
		// Pre-build the map
		m := make(map[int]int, size)
		for i := range size {
			m[i] = i
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				for i := range size {
					_ = m[i]
				}
			}
		})
	}
}

// BenchmarkTrieUpdate measures the cost of updating a single key (immutable - creates new structure)
func BenchmarkTrieUpdate(b *testing.B) {
	for _, size := range sizes {
		// Pre-build the trie
		var root node[int, int]
		for i := range size {
			h := hash(i)
			root = root.insert(i, i, h, 0)
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			h := hash(0)
			for range b.N {
				// Update a single key - this creates a new root with path copying
				_ = root.insert(0, 999, h, 0)
			}
		})
	}
}

// BenchmarkMapUpdate measures the cost of updating a single key (mutable)
func BenchmarkMapUpdate(b *testing.B) {
	for _, size := range sizes {
		// Pre-build the map
		m := make(map[int]int, size)
		for i := range size {
			m[i] = i
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				m[0] = 999
			}
		})
	}
}

// BenchmarkTrieMixedOps simulates realistic usage with mixed reads and writes
func BenchmarkTrieMixedOps(b *testing.B) {
	for _, size := range sizes {
		var root node[int, int]
		for i := range size {
			h := hash(i)
			root = root.insert(i, i, h, 0)
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := range b.N {
				h := hash(i % size)
				if i%10 == 0 {
					// 10% writes
					root = root.insert(i%size, i, h, 0)
				} else {
					// 90% reads
					root.get(i%size, h, 0)
				}
			}
		})
	}
}

// BenchmarkMapMixedOps simulates realistic usage with mixed reads and writes
func BenchmarkMapMixedOps(b *testing.B) {
	for _, size := range sizes {
		m := make(map[int]int, size)
		for i := range size {
			m[i] = i
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := range b.N {
				if i%10 == 0 {
					// 10% writes
					m[i%size] = i
				} else {
					// 90% reads
					_ = m[i%size]
				}
			}
		})
	}
}

// BenchmarkMemoryUsage measures memory consumption per entry
func BenchmarkMemoryUsage(b *testing.B) {
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Trie/size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				var root node[int, int]
				for i := range size {
					h := hash(i)
					root = root.insert(i, i, h, 0)
				}
				// Prevent optimization
				if root.isEmpty() {
					b.Fatal("unexpected empty")
				}
			}
		})

		b.Run(fmt.Sprintf("Builder/size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				builder := NewBuilder[int, int]()
				for i := range size {
					builder.Set(i, i)
				}
				m := builder.Build()
				if m.Len() != size {
					b.Fatal("unexpected length")
				}
			}
		})

		b.Run(fmt.Sprintf("StdMap/size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				m := make(map[int]int)
				for i := range size {
					m[i] = i
				}
				if len(m) != size {
					b.Fatal("unexpected length")
				}
			}
		})
	}
}

// func TestKeyPrefix(t *testing.T) {
// 	// Test with a known key value
// 	// Using 0xFFFFFFFFFFFFFFFF (all 1s) makes it easy to verify masking
// 	allOnes := hashedKey(0xFFFFFFFFFFFFFFFF)

// 	tests := []struct {
// 		name     string
// 		key      hashedKey
// 		depth    uint
// 		expected hashedKey
// 	}{
// 		{
// 			name:     "depth 0, width 4 - should keep top 4 bits",
// 			key:      allOnes,
// 			depth:    0,
// 			expected: 0xF000000000000000, // top 4 bits
// 		},
// 		{
// 			name:     "depth 1, width 4 - should keep top 8 bits",
// 			key:      allOnes,
// 			depth:    1,
// 			expected: 0xFF00000000000000, // top 8 bits
// 		},
// 		{
// 			name:     "depth 2, width 4 - should keep top 12 bits",
// 			key:      allOnes,
// 			depth:    2,
// 			expected: 0xFFF0000000000000, // top 12 bits
// 		},
// 		{
// 			name:     "depth 3, width 4 - should keep top 16 bits",
// 			key:      allOnes,
// 			depth:    3,
// 			expected: 0xFFFF000000000000, // top 16 bits
// 		},
// 		{
// 			name:     "specific key at depth 0",
// 			key:      0x123456789ABCDEF0,
// 			depth:    0,
// 			expected: 0x1000000000000000, // top 4 bits of 0x123...
// 		},
// 		{
// 			name:     "specific key at depth 1",
// 			key:      0x123456789ABCDEF0,
// 			depth:    1,
// 			expected: 0x1200000000000000, // top 8 bits
// 		},
// 		{
// 			name:     "zero key at any depth",
// 			key:      0,
// 			depth:    2,
// 			expected: 0,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			result := hintPrefix(tc.key, tc.depth)
// 			if result != tc.expected {
// 				t.Errorf("keyPrefix(0x%X, %d) = 0x%X, expected 0x%X",
// 					tc.key, tc.depth, result, tc.expected)
// 			}
// 		})
// 	}
// }

// func TestKeyPrefixEdgeCases(t *testing.T) {
// 	// Test that offset calculation doesn't overflow or underflow
// 	t.Run("large depth should not panic", func(t *testing.T) {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				t.Errorf("keyPrefix panicked with large depth: %v", r)
// 			}
// 		}()
// 		// With width=4, depth=4 gives width<<depth = 64, so offset = 0
// 		result := hintPrefix(0xFFFFFFFFFFFFFFFF, 4)
// 		t.Logf("depth=4: result = 0x%X", result)
// 	})

// 	t.Run("depth causing negative offset", func(t *testing.T) {
// 		// With width=4, depth=5 gives width<<depth = 128, so offset = 64-128 = -64
// 		// This would cause undefined behavior with shifts
// 		defer func() {
// 			if r := recover(); r != nil {
// 				t.Logf("keyPrefix panicked as expected with overflow: %v", r)
// 			}
// 		}()
// 		result := hintPrefix(0xFFFFFFFFFFFFFFFF, 5)
// 		t.Logf("depth=5: result = 0x%X (potential overflow issue)", result)
// 	})
// }

// func TestKeyPrefixLinearGrowth(t *testing.T) {
// 	// This test verifies that the prefix grows linearly with depth
// 	// Each depth level should add 'width' more bits to the prefix
// 	allOnes := hashedKey(0xFFFFFFFFFFFFFFFF)
// 	width := uint(4)

// 	var prevBitCount int
// 	for depth := range uint(4) {
// 		result := hintPrefix(allOnes, depth)

// 		// Count how many bits are set (from the top)
// 		bitCount := 0
// 		for i := 63; i >= 0; i-- {
// 			if result&(1<<i) != 0 {
// 				bitCount++
// 			} else {
// 				break
// 			}
// 		}

// 		expectedBits := int(width * (depth + 1))
// 		t.Logf("depth=%d: result=0x%016X, bits set from top=%d, expected=%d",
// 			depth, result, bitCount, expectedBits)

// 		if bitCount != expectedBits {
// 			t.Errorf("At depth %d: got %d prefix bits, expected %d (linear growth)",
// 				depth, bitCount, expectedBits)
// 		}

// 		if depth > 0 && bitCount <= prevBitCount {
// 			t.Errorf("Prefix should grow with depth: depth %d has %d bits, but depth %d had %d bits",
// 				depth, bitCount, depth-1, prevBitCount)
// 		}
// 		prevBitCount = bitCount
// 	}
// }
