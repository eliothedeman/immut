package immut

import "testing"

func TestListAppend(t *testing.T) {
	l := NewList(1)

	if l.Len() != 1 {
		t.Errorf("Expected 1 got %d", l.Len())
	}

	l.Append(100)
	if l.Len() != 1 {
		t.Errorf("Expcected 1 got %d", l.Len())
	}

	if l.Append(4).Len() != 2 {
		t.Errorf("Expected 2 got %d", l.Append(4).Len())
	}

	x := l.Append(4)
	i, err := x.Index(1)
	if err != nil {
		t.Error(err)
	}

	if i != 4 {
		t.Errorf("Expected 4 got %d", i)
	}
}

func TestListPrepend(t *testing.T) {
	l := NewList(1)

	if l.Len() != 1 {
		t.Errorf("Expcted 1 got %d", l.Len())
	}

	l.Prepend(2)

	if l.Len() != 1 {
		t.Errorf("Expected 1 got %d", l.Len())
	}

	x := l.Prepend(2)

	if x.Len() != 2 {
		t.Errorf("Expected 2 got %d", x.Len())
	}

	i, err := x.Index(0)
	if err != nil {
		t.Error(err)
	}

	if i != 2 {
		t.Errorf("Expcted 2 got %d", i)
	}

}
