package store

import (
	"os"
	"time"
	"urllite/types"

	"github.com/gocql/gocql"
)

type store struct {
	DBSession *gocql.Session
}

type Store interface {
	CreateUser(user *types.User) error
}

func NewStore() Store {
	session, err := gocql.NewCluster(os.Getenv("CASSANDRA_HOST")).CreateSession()
	if err != nil {
		panic(err)
	}
	return &store{DBSession: session}
}

func (s *store) CreateUser(user *types.User) error {
	// Implement user creation logic
	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	createUserQuery := `INSERT INTO ` + keyspace + `.users (id, name, email, mobile, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	user.CreatedAt, user.UpdatedAt, user.ID = time.Now(), time.Now(), gocql.TimeUUID() // Generate a new UUID for the user adn set timestamps
	user.Status = "active" // Set default status to active

	return s.DBSession.Query(createUserQuery, user.ID, user.Name, user.Email, user.Mobile, user.Status, user.CreatedAt, user.UpdatedAt).Exec()
}
