package notify

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

func parseClock(s string) (pgtype.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return pgtype.Time{}, err
	}
	return pgtype.Time{Microseconds: int64(t.Hour()*3600+t.Minute()*60) * 1e6, Valid: true}, nil
}
