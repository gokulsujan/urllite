package store

import (
	"fmt"
	"os"
	"time"
	"urllite/types"

	"github.com/gocql/gocql"
)

type store struct {
	DBSession *gocql.Session
}

type Store interface {

	//User store

	CreateUser(user *types.User) error
	GetUserByID(id string) (*types.User, error)
	GetUserByEmail(email string) (*types.User, error)
	SearchUsers(filter types.UserFilter) ([]*types.User, error)
	UpdateUser(user *types.User) error
	DeleteUser(user *types.User) error


	//Password Store
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
	user.Status = "active"                                                             // Set default status to active

	return s.DBSession.Query(createUserQuery, user.ID, user.Name, user.Email, user.Mobile, user.Status, user.CreatedAt, user.UpdatedAt).Exec()
}

func (s *store) GetUserByID(id string) (*types.User, error) {
	var user types.User
	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	getUserQueryById := "SELECT  id, name, email, mobile, status, created_at, updated_at, deleted_at FROM " + keyspace + ".users where id = ?"
	if err := s.DBSession.Query(getUserQueryById, id).Consistency(gocql.One).Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return nil, err
	}

	if !user.DeletedAt.IsZero() {
		return nil, gocql.ErrNotFound
	}

	if user.Name == "" && user.Email == "" && user.Mobile == "" && user.Status == "" && user.CreatedAt.IsZero() && user.UpdatedAt.IsZero() {
		return nil, nil
	}

	return &user, nil
}

func (s *store) GetUserByEmail(email string) (*types.User, error) {
	var user types.User
	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	getUserQuery := `SELECT id, name, email, mobile, status, created_at, updated_at, deleted_at FROM ` + keyspace + `.users WHERE email = ? ALLOW FILTERING`
	if err := s.DBSession.Query(getUserQuery, email).Consistency(gocql.One).Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return nil, err
	}

	isDeleted := !user.DeletedAt.IsZero()

	if !isDeleted && user.Email == "" && user.Mobile == "" && user.Status == "" && user.CreatedAt.IsZero() && user.UpdatedAt.IsZero() {
		return nil, nil
	}
	return &user, nil
}

func (s *store) SearchUsers(filter types.UserFilter) ([]*types.User, error) {
	var users []*types.User
	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	searchUsersQuery := `SELECT id, name, email, mobile, status, created_at, updated_at FROM ` + keyspace + `.users`
	if filter.Name != "" || filter.Email != "" || filter.Mobile != "" || filter.Status != "" {
		searchUsersQuery += ` WHERE`
	}

	// Build the WHERE clause based on the filter
	filterStr := ""
	values := []interface{}{}
	if filter.Name != "" {
		if filterStr != "" {
			filterStr += " AND"
		}
		filterStr += " name = ?"
		values = append(values, filter.Name)
	}

	if filter.Email != "" {
		if filterStr != "" {
			filterStr += " AND"
		}
		filterStr += " email = ?"
		values = append(values, filter.Email)

	}

	if filter.Mobile != "" {
		if filterStr != "" {
			filterStr += " AND"
		}
		filterStr += " mobile = ?"
		values = append(values, filter.Mobile)

	}

	if filter.Status != "" {
		if filterStr != "" {
			filterStr += " AND"
		}
		filterStr += " status = ?"

		values = append(values, filter.Status)

	}
	if filterStr != "" {
		searchUsersQuery += filterStr
	}

	// Add ALLOW FILTERING to the query
	searchUsersQuery += " ALLOW FILTERING"
	// Execute the query
	var iter *gocql.Iter
	if len(values) > 0 {
		iter = s.DBSession.Query(searchUsersQuery, values).Iter()
	} else {
		iter = s.DBSession.Query(searchUsersQuery).Iter()
	}
	defer iter.Close()

	// Iterate over the results
	for {
		var user types.User
		if !iter.Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Status, &user.CreatedAt, &user.UpdatedAt) {
			break
		}
		if user.DeletedAt.IsZero() {
			users = append(users, &user)
		}
	}

	// Check for errors during iteration
	if err := iter.Close(); err != nil {
		return nil, err
	}

	// If no users are found, return nil
	if len(users) == 0 {
		return nil, nil
	}

	// Return the list of users
	return users, nil

}

func (s *store) UpdateUser(user *types.User) error {
	if user.ID == (gocql.UUID{}) {
		return fmt.Errorf("No user id found")
	}

	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	userUpdateQuery := "UPDATE " + keyspace + ".users SET name = ?, email = ?, mobile = ?, verified_email = ?, status = ?, updated_at = ? WHERE id = ?"
	return s.DBSession.Query(userUpdateQuery, user.Name, user.Email, user.Mobile, user.VerifiedEmail, user.Status, time.Now(), user.ID).Exec()
}

func (s *store) DeleteUser(user *types.User) error {
	if user.ID == (gocql.UUID{}) {
		return fmt.Errorf("No user id found")
	}

	keyspace := os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	userDeleteQuery := "UPDATE " + keyspace + ".users SET deleted_at = ? WHERE id = ?"

	return s.DBSession.Query(userDeleteQuery, time.Now(), user.ID.String()).Exec()
}
