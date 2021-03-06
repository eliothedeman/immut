package immut

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	IndexOutOfRange = errors.New("index out of range")
)

// A List is an immutable singly linked list that is safe for concurrent use
type List struct {
	next *List
	val  interface{}
}

// NewList creates and returns an new list with the given value at the first node
func NewList(val interface{}) *List {
	return &List{
		val: val,
	}
}

// Val returns the value stored at the current node in the list
func (l *List) Val() interface{} {
	return l.val
}

// Len returns the length of the list
func (l *List) Len() int {
	i := 1
	y := l
	for !y.End() {
		i++
		y = y.next
	}

	return i
}

// String returns a string representation of the list
func (l *List) String() string {
	if l == nil {
		return "nil"
	}
	b := bytes.NewBuffer(nil)
	b.WriteString("[")
	y := l
	for {
		b.WriteString(fmt.Sprintf("%v", y.val))
		if !y.End() {
			b.WriteString(", ")
		} else {
			break
		}
		y = y.next
	}
	b.WriteString("]")

	return b.String()
}

// End returns true if this is the end of the list
func (l *List) End() bool {
	return l.next == nil
}

// Index returns the value stored at the given index if it exists
func (l *List) Index(i int) (interface{}, error) {
	x := 0
	y := l

	for x < i {

		if y.End() {
			return nil, IndexOutOfRange
		}
		y = y.next
		x++
	}

	return y.val, nil
}

// Prepend the given value onto a new list
func (l *List) Prepend(val interface{}) *List {
	return &List{
		next: l,
		val:  val,
	}
}

// Append the given value to the end of the list. This will reallocate the whole list
func (l *List) Append(val interface{}) *List {

	// make a copy of this list
	n := &List{}
	n.val = l.val

	//  if this is not the end, pass it down the line
	if !l.End() {
		n.next = l.next.Append(val)
	} else {
		n.next = &List{
			val: val,
		}
	}

	return n
}

// Next returns the next node in the list
func (l *List) Next() *List {
	return l.next
}

func (l *List) Each(f func(i interface{})) {
	if l == nil {
		return
	}

	f(l.val)
	l.Next().Each(f)
}

func (l *List) Filter(f func(*List) bool) *List {
	if l == nil {
		return nil
	}

	if f(l) {

		n := NewList(l.val)
		n.next = n.next.Filter(f)
		return n
	}

	if l.End() {
		return nil
	}

	return l.next.Filter(f)

}
