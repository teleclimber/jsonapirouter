package jsonapirouter

import "testing"

func TestGetIndex(t *testing.T) {
	incs := NewIncludes(nil)

	// the first one should be zero
	i := incs.getIndex("abc", "0")
	if i != 0 {
		t.Error("expected 0")
	}

	// get it again it should be zero aagain
	i = incs.getIndex("abc", "0")
	if i != 0 {
		t.Error("expected 0")
	}

	// get another, it should be 1
	i = incs.getIndex("abc", "uno")
	if i != 1 {
		t.Error("expected 1")
	}

	// get it again, it should be 1
	i = incs.getIndex("abc", "uno")
	if i != 1 {
		t.Error("expected 1")
	}
}
