package storage

import "testing"

func TestMemoryCursorEmpty(t *testing.T) {
	s := NewMemoryCursorStore()

	got, err := s.GetCursor("horizon-operations")

	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestMemoryCursorRoundTrip(t *testing.T) {
	s := NewMemoryCursorStore()

	if err := s.SetCursor("horizon-operations", "pt-9"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetCursor("horizon-operations")

	if err != nil {
		t.Fatal(err)
	}
	if got != "pt-9" {
		t.Fatalf("got %q", got)
	}
}
