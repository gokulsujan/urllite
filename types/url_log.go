package types

import (
	"time"

	"github.com/gocql/gocql"
)

type UrlLog struct {
	ID             gocql.UUID `json:"id"`
	UrlID          gocql.UUID `json:"url_id`
	VisitedAt      time.Time  `json:"visited_at"`
	RedirectStatus string     `json:"redirect_status"`
	HttpStatusCode int        `json:"http_status_code"`
	ClientIP       string     `json:"client_ip"`
	City           string     `json:"city"`
	Region         string     `json:"region"`
	Isp            string     `json:"isp"`
	Timezone       string     `json:"timezone"`
	Country        string     `json:"country"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
