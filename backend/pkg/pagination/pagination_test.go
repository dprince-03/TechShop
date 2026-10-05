package pagination

import (
	"errors"
	"testing"
	"time"
)

func TestCursorRoundTripAndTamper(t *testing.T) {
	c := NewCursors("test-key-for-signing-cursors")
	at := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	page := Paginate(c, []time.Time{at.Add(time.Hour), at, at.Add(-time.Hour)}, 2, func(v time.Time) time.Time { return v })
	if !page.Page.HasMore || len(page.Items) != 2 {
		t.Fatalf("want 2 items and more, got %d items, hasMore=%v", len(page.Items), page.Page.HasMore)
	}
	got, err := c.Before(page.Page.NextCursor)
	if err != nil || !got.Equal(at) {
		t.Fatalf("round trip: got %v, %v", got, err)
	}
	if _, err := NewCursors("another-key").Before(page.Page.NextCursor); !errors.Is(err, ErrBadCursor) {
		t.Fatalf("cursor signed with another key must be rejected, got %v", err)
	}
}
