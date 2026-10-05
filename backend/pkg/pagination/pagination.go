// Package pagination provides signed, tamper-proof cursors for keyset pagination.
package pagination

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// MaxPageSize caps every list endpoint.
const MaxPageSize = 100

// Page is the pagination block of every list response.
type Page struct {
	NextCursor string `json:"nextCursor,omitempty" doc:"Opaque cursor for the next page"`
	HasMore    bool   `json:"hasMore"`
}

// Cursors signs and verifies opaque pagination cursors so clients can't forge offsets.
type Cursors struct{ key []byte }

// NewCursors creates a cursor codec.
func NewCursors(key string) *Cursors { return &Cursors{key: []byte(key)} }

// Encode turns any JSON-able position into a signed cursor.
func (c *Cursors) Encode(v any) string {
	b, _ := json.Marshal(v)
	m := hmac.New(sha256.New, c.key)
	m.Write(b)
	return base64.RawURLEncoding.EncodeToString(b) + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil)[:12])
}

// Decode verifies and decodes a cursor into v. An empty cursor leaves v untouched.
func (c *Cursors) Decode(cursor string, v any) error {
	if cursor == "" {
		return nil
	}
	body, sig, ok := strings.Cut(cursor, ".")
	if !ok {
		return errors.New("bad cursor")
	}
	b, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return errors.New("bad cursor")
	}
	m := hmac.New(sha256.New, c.key)
	m.Write(b)
	want := base64.RawURLEncoding.EncodeToString(m.Sum(nil)[:12])
	if !hmac.Equal([]byte(want), []byte(sig)) {
		return errors.New("bad cursor")
	}
	return json.Unmarshal(b, v)
}

// Limit clamps a requested page size to 1..MaxPageSize (default 20).
func Limit(n int) int {
	switch {
	case n <= 0:
		return 20
	case n > MaxPageSize:
		return MaxPageSize
	}
	return n
}

// ErrBadCursor is returned when a cursor fails verification.
var ErrBadCursor = errors.New("pagination: invalid cursor")

// List is a page of items plus its pagination block.
type List[T any] struct {
	Items []T  `json:"items"`
	Page  Page `json:"page"`
}

// Paginate trims rows fetched with limit+1 and builds the next cursor from the last kept row.
func Paginate[T any](c *Cursors, rows []T, limit int, key func(T) time.Time) List[T] {
	if rows == nil {
		rows = []T{}
	}
	if len(rows) <= limit {
		return List[T]{Items: rows}
	}
	rows = rows[:limit]
	return List[T]{Items: rows, Page: Page{HasMore: true, NextCursor: c.Encode(struct{ Before time.Time }{key(rows[limit-1])})}}
}

// Before decodes a cursor made by Paginate (nil when the cursor is empty).
func (c *Cursors) Before(cursor string) (*time.Time, error) {
	if cursor == "" {
		return nil, nil
	}
	var v struct{ Before time.Time }
	if err := c.Decode(cursor, &v); err != nil {
		return nil, ErrBadCursor
	}
	return &v.Before, nil
}
