# Go Testing

- `testing` + `testify/require` (or stdlib only if the project prefers).
- Table-driven tests:
```go
func TestCalculateTotal(t *testing.T) {
    tests := []struct{ name string; items []Item; want int64; wantErr error }{
        {"empty cart", nil, 0, nil},
        {"negative qty", []Item{{Qty: -1}}, 0, ErrInvalidQty},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateTotal(tt.items)
            require.ErrorIs(t, err, tt.wantErr)
            require.Equal(t, tt.want, got)
        })
    }
}
```
- HTTP: `httptest.NewRecorder()` with the real Gin router.
- DB integration: `testcontainers-go` Postgres/Mongo/Redis; run migrations once per package (`TestMain`).
- Mocks: hand-written fakes for small interfaces; `mockery`/`gomock` if project uses them.
- Run with `go test ./... -race -cover`; use `t.Parallel()` where safe.
- Fuzz critical parsers: `func FuzzParse(f *testing.F)`.
