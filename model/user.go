package model

import (
	"time"

	"github.com/gocql/gocql"
)

type User struct {
	ID            gocql.UUID `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	VerifiedEmail string     `json:"-"` // json ignore field
	Mobile        string     `json:"mobile"`
	Status        string     `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
