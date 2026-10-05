package httpx

import (
	"time"

	"github.com/dprince-03/techshop/backend/pkg/pagination"
)

// Re-exported so handlers keep one import; the logic lives in pkg/pagination.
const MaxPageSize = pagination.MaxPageSize

type (
	Page        = pagination.Page
	List[T any] = pagination.List[T]
)

// Cursors wraps pagination.Cursors so a bad cursor becomes a 422 problem response.
type Cursors struct{ *pagination.Cursors }

// NewCursors signs cursors with the given key.
func NewCursors(key string) *Cursors { return &Cursors{pagination.NewCursors(key)} }

// Limit clamps a requested page size.
func Limit(n int) int { return pagination.Limit(n) }

// Paginate builds a page from rows fetched with limit+1.
func Paginate[T any](c *Cursors, rows []T, limit int, key func(T) time.Time) List[T] {
	return pagination.Paginate(c.Cursors, rows, limit, key)
}

// Before decodes a cursor made by Paginate, as a 422 problem when it is invalid.
func (c *Cursors) Before(cursor string) (*time.Time, error) {
	t, err := c.Cursors.Before(cursor)
	if err != nil {
		return nil, BadCursor()
	}
	return t, nil
}

// BadCursor is returned when a cursor fails verification.
func BadCursor() *Problem {
	return Invalid("bad_cursor", "The page cursor is invalid; start again from the first page.")
}
