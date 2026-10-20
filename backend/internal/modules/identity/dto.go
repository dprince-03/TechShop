package identity

import (
	"time"

	"github.com/google/uuid"
)

type tokensOut struct{ Body *TokenPair }

type userSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty" doc:"Masked"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}
