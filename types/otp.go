package types

import (
	"time"

	"github.com/gocql/gocql"
)

type Otp struct {
	ID        gocql.UUID `json:"id"`
	UserID    gocql.UUID `json:"user_id"`
	Key       string     `json:"key"`
	Otp       string     `json:"otp"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiredAt time.Time  `json:"expired_at"`
}
