package model

import (
	"time"

	"github.com/gocql/gocql"
)

type URL struct {
	ID       gocql.UUID `json:"id"`
	UserID   gocql.UUID `json:"user_id"`
	LongUrl  string     `josn:"long_url"`
	ShortUrl string     `json:"short_url"`
	Status   string     `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
