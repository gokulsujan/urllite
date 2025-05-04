package store

import (
	"fmt"
	"os"
	"time"
	"urllite/types"
	"urllite/types/dtos"

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
	UserDashboardStats(userID string) (*dtos.AdminUserDashboardDTO, error)

	//Password Store
	CreatePassword(password *types.Password) error
	GetPasswordByUserID(userID string) (*types.Password, error)
	UpdatePassword(passwrod *types.Password) error
	DeletePassword(password *types.Password) error

	//URL Store
	CreateURL(url *types.URL) error
	GetUrlByID(id string) (*types.URL, error)
	GetUrlByShortUrl(short_url string) (*types.URL, error)
	GetURLsOfUser(user_id string) ([]*types.URL, error)
	DeleteURL(url *types.URL) error

	//URL Logs
	CreateUrlLog(log *types.UrlLog) error
	DeleteUrlLogsByUrlId(urlID string, deletedTime time.Time) error
	GetUrlLogsByUrlId(urlID string) ([]*types.UrlLog, error)
	CountInteractions(urlId string) (int, error)

	// OTP
	CreateOtp(otp *types.Otp) (*types.Otp, error)
	GetOtpByUserIdAndOtp(userId, key, otpValue string) ([]*types.Otp, error)
	ChangeOtpStatus(otp *types.Otp, status string) error

	//Admin
	AdminDashboard() (*dtos.AdminDashboardDTO, error)
}

var CASSANDRA_HOST, CASSANDRA_KEYSPACE string

func NewStore() Store {
	CASSANDRA_HOST = os.Getenv("CASSANDRA_HOST")
	CASSANDRA_KEYSPACE = os.Getenv("CASSANDRA_URLLITE_KEYSPACE")
	session, err := gocql.NewCluster(CASSANDRA_HOST).CreateSession()
	if err != nil {
		fmt.Println("Host: " + CASSANDRA_HOST)
		fmt.Println("Keyspave: " + CASSANDRA_KEYSPACE)
		fmt.Println("Error: " + err.Error())
		panic(err)
	}
	return &store{DBSession: session}
}

func (s *store) CreateUser(user *types.User) error {
	createUserQuery := `INSERT INTO ` + CASSANDRA_KEYSPACE + `.users (id, name, email, mobile, status, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	user.CreatedAt, user.UpdatedAt, user.ID = time.Now(), time.Now(), gocql.TimeUUID() // Generate a new UUID for the user adn set timestamps
	user.Status, user.Role = "active", "user"
	return s.DBSession.Query(createUserQuery, user.ID, user.Name, user.Email, user.Mobile, user.Status, user.Role, user.CreatedAt, user.UpdatedAt).Exec()
}

func (s *store) GetUserByID(id string) (*types.User, error) {
	var user types.User
	getUserQueryById := "SELECT  id, name, email, mobile, verified_email, status, role, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".users where id = ?"
	if err := s.DBSession.Query(getUserQueryById, id).Consistency(gocql.One).Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.VerifiedEmail, &user.Status, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
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
	getUserQuery := `SELECT id, name, email, mobile, verified_email, status, role, created_at, updated_at, deleted_at FROM ` + CASSANDRA_KEYSPACE + `.users WHERE email = ? ALLOW FILTERING`
	if err := s.DBSession.Query(getUserQuery, email).Consistency(gocql.One).Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.VerifiedEmail, &user.Status, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
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
	searchUsersQuery := `SELECT id, name, email, mobile, verified_email, status, role, created_at, updated_at, deleted_at FROM ` + CASSANDRA_KEYSPACE + `.users`
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
		if !iter.Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.VerifiedEmail, &user.Status, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt) {
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

	userUpdateQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".users SET name = ?, email = ?, mobile = ?, verified_email = ?, status = ?, role =?, updated_at = ? WHERE id = ?"
	return s.DBSession.Query(userUpdateQuery, user.Name, user.Email, user.Mobile, user.VerifiedEmail, user.Status, user.Role, time.Now(), user.ID).Exec()
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

func (s *store) GetUrlByShortUrl(short_url string) (*types.URL, error) {
	var url types.URL
	selectUrlByIdQuery := "SELECT id, user_id, long_url, short_url, status, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".urls WHERE short_url = ? ALLOW FILTERING"
	err := s.DBSession.Query(selectUrlByIdQuery, short_url).Consistency(gocql.One).Scan(&url.ID, &url.UserID, &url.LongUrl, &url.ShortUrl, &url.Status, &url.CreatedAt, &url.UpdatedAt, &url.DeletedAt)
	if !url.DeletedAt.IsZero() {
		return nil, nil
	}

	if err == gocql.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &url, nil
}

func (s *store) GetURLsOfUser(user_id string) ([]*types.URL, error) {
	var urls []*types.URL
	getURLsQuery := "SELECT id, user_id, long_url, short_url, status, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".urls WHERE user_id = ? ALLOW FILTERING"
	iter := s.DBSession.Query(getURLsQuery, user_id).Iter()

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

func (s *store) CreateUrlLog(log *types.UrlLog) error {
	insertUrlLogQuery := "INSERT INTO " + CASSANDRA_KEYSPACE + ".url_logs (id, url_id, visited_at, redirect_status, http_status_code, client_ip, city, country, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	log.ID, log.CreatedAt, log.UpdatedAt = gocql.TimeUUID(), time.Now(), time.Now()
	return s.DBSession.Query(insertUrlLogQuery, log.ID, log.UrlID, log.VisitedAt, log.RedirectStatus, log.HttpStatusCode, log.ClientIP, log.City, log.Country, log.CreatedAt, log.UpdatedAt).Exec()
}

func (s *store) GetUrlLogsByUrlId(urlID string) ([]*types.UrlLog, error) {
	searchLogsQuery := "SELECT id, client_ip, city, country, url_id, visited_at, redirect_status, http_status_code, created_at, updated_at, deleted_at FROM " + CASSANDRA_KEYSPACE + ".url_logs WHERE url_id = ? ORDER BY created_at DESC"
	url, err := s.GetUrlByID(urlID)
	if err != nil {
		return nil, err
	}

	iter := s.DBSession.Query(searchLogsQuery, url.ID).Iter()
	var logs []*types.UrlLog

	for {
		var log types.UrlLog
		if !iter.Scan(&log.ID, &log.ClientIP, &log.City, &log.Country, &log.UrlID, &log.VisitedAt, &log.RedirectStatus, &log.HttpStatusCode, &log.CreatedAt, &log.UpdatedAt, &log.DeletedAt) {
			break
		}
		if log.DeletedAt.IsZero() {
			logs = append(logs, &log)
		}
	}

	if len(logs) == 0 {
		return nil, nil
	}

	return logs, nil

}

func (s *store) DeleteUrlLogsByUrlId(urlID string, deletedTime time.Time) error {
	url, err := s.GetUrlByID(urlID)
	if err != nil {
		return err
	}
	if url == nil {
		return nil
	}
	deleteUrlLogQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".url_logs SET deleted_at = ? WHERE url_id = ?"
	logs, err := s.GetUrlLogsByUrlId(urlID)
	if err != nil {
		return err
	}
	for _, log := range logs {
		s.DBSession.Query(deleteUrlLogQuery, deletedTime, log.ID).Exec()
	}
	return nil
}

func (s *store) CreateOtp(otp *types.Otp) (*types.Otp, error) {
	otpInsertQuery := "INSERT INTO " + CASSANDRA_KEYSPACE + ".otp (id, user_id, key, otp, status, created_at, expired_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	otp.ID = gocql.TimeUUID()
	otp.CreatedAt = time.Now()
	err := s.DBSession.Query(otpInsertQuery, otp.ID, otp.UserID, otp.Key, otp.Otp, otp.Status, otp.CreatedAt, otp.ExpiredAt).Exec()
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (s *store) GetOtpByUserId(userId string) (*types.Otp, error) {
	var otp *types.Otp
	otpGetQuery := "SELECT id, user_id, key, otp, created_at, expired_at FROM " + CASSANDRA_KEYSPACE + ".otp WHERE user_id =?"
	err := s.DBSession.Query(otpGetQuery, userId).Consistency(gocql.One).Scan(otp.ID, otp.UserID, otp.Otp, otp.CreatedAt, otp.ExpiredAt)
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (s *store) GetOtpByUserIdAndOtp(userId, key, otpStr string) ([]*types.Otp, error) {
	var otps []*types.Otp
	otpGetQuery := "SELECT id, user_id, key, otp, status, created_at, expired_at FROM " + CASSANDRA_KEYSPACE + ".otp WHERE user_id =? AND expired_at > toTimestamp(now()) AND key = ? AND otp = ? ALLOW FILTERING"
	iter := s.DBSession.Query(otpGetQuery, userId, key, otpStr).Iter()
	defer iter.Close()
	for {
		var otp types.Otp
		if !iter.Scan(&otp.ID, &otp.UserID, &otp.Key, &otp.Otp, &otp.Status, &otp.CreatedAt, &otp.ExpiredAt) {
			scannerErr := iter.Scanner().Err()
			if scannerErr != nil {
				return nil, scannerErr
			}
			break
		}

		if otp.Status == "pending" {
			otps = append(otps, &otp)
		}
	}

	if len(otps) == 0 {
		return nil, nil
	}

	return otps, nil
}

func (s *store) ChangeOtpStatus(otp *types.Otp, status string) error {
	updateOtpQuery := "UPDATE " + CASSANDRA_KEYSPACE + ".otp SET status = ? WHERE id = ?"
	return s.DBSession.Query(updateOtpQuery, status, otp.ID).Exec()
}

func (s *store) CountInteractions(urlId string) (int, error) {
	var count int
	countInteractonsQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".url_logs WHERE url_id =? ALLOW FILTERING"
	err := s.DBSession.Query(countInteractonsQuery, urlId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *store) AdminDashboard() (*dtos.AdminDashboardDTO, error) {
	totalUrlsQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".urls"
	totalDeletedUrlsQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".urls where deleted_at>0 ALLOW FILTERING"

	totalUsersQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".users"
	totalDeletedUsersQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".users where deleted_at>0 ALLOW FILTERING"
	totalSuspendedUsersQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".users where status='suspended' ALLOW FILTERING"

	var totalUrls, totalDeletedUrls, totalUsers, totalDeletedUsers, totalSuspendedUsers int
	err := s.DBSession.Query(totalUrlsQuery).Scan(&totalUrls)
	if err != nil {
		return nil, err
	}
	err = s.DBSession.Query(totalDeletedUrlsQuery).Scan(&totalDeletedUrls)
	if err != nil {
		return nil, err
	}
	err = s.DBSession.Query(totalUsersQuery).Scan(&totalUsers)
	if err != nil {
		return nil, err
	}
	err = s.DBSession.Query(totalDeletedUsersQuery).Scan(&totalDeletedUsers)
	if err != nil {
		return nil, err
	}
	err = s.DBSession.Query(totalSuspendedUsersQuery).Scan(&totalSuspendedUsers)
	if err != nil {
		return nil, err
	}

	return &dtos.AdminDashboardDTO{
		TotalActiveUrls:             totalUrls - totalDeletedUrls,
		TotalActiveUsers:            totalUsers - totalDeletedUsers - totalSuspendedUsers,
		TotalUsers:                  totalUsers - totalDeletedUsers,
		TotalSuspendedUsers:         totalSuspendedUsers,
		TotalActiveCustomDomains:    0, // Custom domain set up not yet released
		TotalActiveCustomDomainUrls: 0,
	}, nil
}

func (s *store) UserDashboardStats(userID string) (*dtos.AdminUserDashboardDTO, error) {
	totalUrlsQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".urls where user_id = ? ALLOW FILTERING"
	totalDeletedUrlsQuery := "SELECT COUNT(*) FROM " + CASSANDRA_KEYSPACE + ".urls where deleted_at> 0 AND user_id = ?  ALLOW FILTERING"

	var totalUrls, totalDeletedUrls int
	err := s.DBSession.Query(totalUrlsQuery, userID).Scan(&totalUrls)
	if err != nil {
		return nil, err
	}
	err = s.DBSession.Query(totalDeletedUrlsQuery, userID).Scan(&totalDeletedUrls)
	if err != nil {
		return nil, err
	}

	return &dtos.AdminUserDashboardDTO{
		TotalActiveUrls:             totalUrls - totalDeletedUrls,
		TotalActiveCustomDomains:    0, // Custom domain set up not yet released
		TotalActiveCustomDomainUrls: 0,
	}, nil
}
