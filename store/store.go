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
	CreatePassword(password *types.Password) error
	GetPasswordByUserID(userID string) (*types.Password, error)
	UpdatePassword(passwrod *types.Password) error
	DeletePassword(password *types.Password) error

	//URL Store
	CreateURL(url *types.URL) error
	GetUrlByID(urlID string) (*types.URL, error)
	GetURLs() ([]*types.URL, error)
	DeleteURL(url *types.URL) error
}

var CASSANDRA_HOST, CASSANDRA_KEYSPACE string

func NewStore() Store {
	CASSANDRA_HOST = os.Getenv("CASSANDRA_HOST")
	CASSANDRA_KEYSPACE = os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	session, err := gocql.NewCluster(CASSANDRA_HOST).CreateSession()
	if err != nil {
		panic(err)
	}
	return &store{DBSession: session}
}

func (s *store) CreateUser(user *types.User) error {
	// Implement user creation logic
	createUserQuery := `INSERT INTO ` + CASSANDRA_KEYSPACE + `.users (id, name, email, mobile, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	user.CreatedAt, user.UpdatedAt, user.ID = time.Now(), time.Now(), gocql.TimeUUID() // Generate a new UUID for the user adn set timestamps
	user.Status = "active"                                                             // Set default status to active

	return s.DBSession.Query(createUserQuery, user.ID, user.Name, user.Email, user.Mobile, user.Status, user.CreatedAt, user.UpdatedAt).Exec()
}

func (s *store) GetUserByID(id string) (*types.User, error) {
	var user types.User
	getUserQueryById := "SELECT  id, name, email, mobile, status, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".users where id = ?"
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
	getUserQuery := `SELECT id, name, email, mobile, status, created_at, updated_at, deleted_at FROM ` + CASSANDRA_KEYSPACE + `.users WHERE email = ? ALLOW FILTERING`
	if err := s.DBSession.Query(getUserQuery, email).Consistency(gocql.One).Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return nil, err
	}

	if !user.DeletedAt.IsZero() {
		return nil, gocql.ErrNotFound
	}

	if user.Email == "" && user.Mobile == "" && user.Status == "" && user.CreatedAt.IsZero() && user.UpdatedAt.IsZero() {
		return nil, nil
	}
	return &user, nil
}

func (s *store) SearchUsers(filter types.UserFilter) ([]*types.User, error) {
	var users []*types.User
	searchUsersQuery := `SELECT id, name, email, mobile, status, created_at, updated_at, deleted_at FROM ` + CASSANDRA_KEYSPACE + `.users`
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
		if !iter.Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt) {
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

	userUpdateQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".users SET name = ?, email = ?, mobile = ?, verified_email = ?, status = ?, updated_at = ? WHERE id = ?"
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

func (s *store) CreatePassword(password *types.Password) error {
	_, err := s.GetPasswordByUserID(password.UserID.String())
	if err != gocql.ErrNotFound {
		return fmt.Errorf("Password already setted for the user_id: ", password.UserID)
	} else if err != gocql.ErrNotFound && err != nil {
		return err
	}

	createPasswordQuery := "INSERT INTO " + CASSANDRA_KEYSPACE + ".passwords (id, user_id, hashed_password, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	password.ID, password.CreatedAt, password.UpdatedAt = gocql.TimeUUID(), time.Now(), time.Now()
	return s.DBSession.Query(createPasswordQuery, password.ID, password.UserID, password.HashedPassword, password.Status, password.CreatedAt, password.UpdatedAt).Exec()
}

func (s *store) GetPasswordByUserID(userID string) (*types.Password, error) {
	var password types.Password
	searchPasswordByUserIdQuery := "SELECT id, user_id, hashed_password, status, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".passwords WHERE user_id = ? ALLOW FILTERING"
	userUUID, err := gocql.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	if err := s.DBSession.Query(searchPasswordByUserIdQuery, userUUID).Consistency(gocql.One).Scan(&password.ID, &password.UserID, &password.HashedPassword, &password.Status, &password.CreatedAt, &password.UpdatedAt, &password.DeletedAt); err != nil {
		return nil, err
	}

	if !password.DeletedAt.IsZero() {
		return nil, gocql.ErrNotFound
	}

	return &password, nil
}

func (s *store) UpdatePassword(passwrod *types.Password) error {
	updatePasswordQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".passwords SET hashed_password = ?, updated_at = ? WHERE id = ?"
	return s.DBSession.Query(updatePasswordQuery, passwrod.HashedPassword, time.Now(), passwrod.ID).Exec()
}

func (s *store) DeletePassword(password *types.Password) error {
	deletePasswordQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".passwords SET deleted_at = ? WHERE id = ?"
	return s.DBSession.Query(deletePasswordQuery, time.Now(), password.ID).Exec()
}

func (s *store) CreateURL(url *types.URL) error {
	createUrlQuery := "INSERT INTO " + CASSANDRA_KEYSPACE + ".urls (id, user_id, long_url, short_url, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	url.ID, url.CreatedAt, url.UpdatedAt = gocql.TimeUUID(), time.Now(), time.Now()
	return s.DBSession.Query(createUrlQuery, url.ID, url.UserID, url.LongUrl, url.ShortUrl, url.Status, url.CreatedAt, url.UpdatedAt).Exec()
}

func (s *store) GetUrlByID(id string) (*types.URL, error) {
	var url types.URL
	selectUrlByIdQuery := "SELECT id, user_id, long_url, short_url, status, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".urls WHERE id = ?"
	err := s.DBSession.Query(selectUrlByIdQuery, id).Consistency(gocql.One).Scan(&url.ID, &url.UserID, &url.LongUrl, &url.ShortUrl, &url.Status, &url.CreatedAt, &url.UpdatedAt, &url.DeletedAt)

	if err == gocql.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &url, nil
}

func (s *store) GetURLs() ([]*types.URL, error) {
	var urls []*types.URL
	getURLsQuery := "SELECT id, user_id, long_url, short_url, status, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".urls"
	iter := s.DBSession.Query(getURLsQuery).Iter()

	defer iter.Close()

	// Iterate over the results
	for {
		var url types.URL
		if !iter.Scan(&url.ID, &url.UserID, &url.LongUrl, &url.ShortUrl, &url.Status, &url.CreatedAt, &url.UpdatedAt, &url.DeletedAt) {
			break
		}
		if url.DeletedAt.IsZero() {
			urls = append(urls, &url)
		}
	}

	if len(urls) == 0 {
		return nil, nil
	}

	return urls, nil

}

func (s *store) DeleteURL(url *types.URL) error {
	deleteUrlQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".urls SET deleted_at = ? WHERE id = ?"
	return s.DBSession.Query(deleteUrlQuery, time.Now(), url.ID).Exec()
}
