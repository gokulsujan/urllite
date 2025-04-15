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

	// CreateUser creates a new user in the database.
	// It takes a pointer to a User struct as input and returns an error if any occurs.
	// The function uses a Cassandra query to insert the user details into the database.
	// The user ID is generated using gocql.TimeUUID() and the created_at and updated_at timestamps are set to the current time.
	// The function also sets the default status of the user to "active".
	// If the user already exists, it returns an error.
	CreateUser(user *types.User) error

	// GetUserByID retrieves a user by their ID.
	// It returns the user if found, or an error if any occurs.
	// If the user is not found, it returns nil.
	// If the user is found but marked as deleted, it returns nil.
	// The function uses a Cassandra query to fetch the user details based on the ID.
	// The ID is expected to be in the format of a UUID.
	GetUserByID(id string) (*types.User, error)

	// GetUserByEmail retrieves a user by their email address.
	// It returns the user if found, or an error if any occurs.
	// If the user is not found, it returns nil.
	// If the user is found but marked as deleted, it returns nil.
	// The function uses a Cassandra query to fetch the user details based on the email address.
	GetUserByEmail(email string) (*types.User, error)

	// SearchUsers retrieves a list of users based on the provided filter criteria.
	// It returns a slice of pointers to User structs and an error if any occurs.
	// The filter criteria can include name, email, mobile number, and status.
	SearchUsers(filter types.UserFilter) ([]*types.User, error)

	// Update Users with the given params
	// Returns error if the database update failed
	UpdateUser(user *types.User) error

	DeleteUser(user *types.User) error
}

// NewStore initializes a new store instance with a Cassandra session.
// It reads the Cassandra host from the environment variable "CASSANDRA_HOST".
// If the session cannot be created, it panics.
// The function returns a pointer to the store instance.
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
