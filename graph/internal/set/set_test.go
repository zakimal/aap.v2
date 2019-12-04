package set

import "testing"

// TODO: write set test

// For test
type vertex int64

func (v vertex) ID() int64 {
	return int64(v)
}

func TestInt64SetSame(t *testing.T) {
	var (
		a = make(Int64Set)
		b = make(Int64Set)
		c = a
	)

	if int64SetSame(a, b) {
		t.Error("Independently created sets are expected to be not same")
	}

	if !int64SetSame(a, c) {
		t.Error("Original set and copied one are expected to be same")
	}

	a.Add(1)

	if !int64SetSame(a, c) {
		t.Error("Original and copied one are expected to be not same after addition to original one")
	}

	if !int64SetSame(nil, nil) {
		t.Error("nil sets are expected to be same")
	}

	if int64SetSame(b, nil) {
		t.Error("nil set and empty set are expected to be not same")
	}
}

func TestInt64Set_Add(t *testing.T) {
	s := make(Int64Set)

	if s == nil {
		t.Fatal("Set cannot be created successfully")
	}

	if s.Count() != 0 {
		t.Error("Just created set should not contain any elements")
	}

	s.Add(1)
	s.Add(3)
	s.Add(5)

	if s.Count() != 3 {
		t.Error("Incorrect number of elements are in set after adding")
	}

	if !s.Has(1) || !s.Has(3) || !s.Has(5) {
		t.Error("Set doesn't contain element that was added")
	}

	s.Add(1)

	if s.Count() > 3 {
		t.Error("Set double-adds element (element not unique)")
	} else if s.Count() < 3 {
		t.Error("Set double-add lowered len")
	}

	if !s.Has(1) {
		t.Error("Set doesn't contain double-added element")
	}

	if !s.Has(3) || !s.Has(5) {
		t.Error("Set removes element on double-add")
	}
}

func TestInt64Set_Remove(t *testing.T) {
	s := make(Int64Set)

	s.Add(1)
	s.Add(3)
	s.Add(5)

	s.Remove(1)

	if s.Count() != 2 {
		t.Error("Incorrect number of elements are in set after removing an element")
	}

	if s.Has(1) {
		t.Error("Removed element is in set after removal of it")
	}

	if !s.Has(3) || !s.Has(5) {
		t.Error("Set has removed wrong element")
	}

	s.Remove(1)

	if s.Count() != 2 || s.Has(1) {
		t.Error("Doubly-remove dose something wrong")
	}

	s.Add(1)

	if s.Count() != 3 || !s.Has(1) {
		t.Error("Cannot add element after removal for some reason")
	}
}
