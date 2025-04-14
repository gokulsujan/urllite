package model

import (
	"time"

	"github.com/gocql/gocql"
)

type Password struct {
	ID gocql.UUID `json:"id"`
	UserID gocql.UUID `json:"user_id"`
	HashedPassword string `json:"hashed_password"`
	Status string `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
